package mod

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type Manager struct {
	mods              map[string]Info
	builtinBinaryPath string
}

type Info struct {
	Name       string
	Image      string
	IsBuiltin  bool
	BinaryPath string
	Args       []string
}

type ThirdPartyConfig struct {
	Plugins []ThirdPartyPlugin `json:"plugins"`
}

type ThirdPartyPlugin struct {
	Name string `json:"name"`
}

type ManagerOption func(m *Manager)

func SetBuiltinBinaryPathOption(path string) ManagerOption {
	return func(m *Manager) {
		m.builtinBinaryPath = path
	}
}

func thirdpartyModDir() string {
	if dir, ok := os.LookupEnv("RIG_THIRDPARTY_MOD_DIR"); ok {
		return dir
	}
	if dir, ok := os.LookupEnv("RIG_THIRDPARTY_PLUGIN_DIR"); ok {
		return dir
	}
	return "/app/bin/plugins-thirdparty"
}

func validateModName(s string) error {
	i := strings.Index(s, ".")
	if i == -1 {
		return fmt.Errorf(
			"mod name '%s' was malformed. Must be of the form <group>.<name> where <group> and <name> are system names",
			s,
		)
	}

	group, name := s[:i], s[i+1:]
	if err := common.ValidateKubernetesName(group); err != nil {
		return fmt.Errorf("mod name '%s' was malformed: %q", s, err)
	}
	if err := common.ValidateKubernetesName(name); err != nil {
		return fmt.Errorf("mod name '%s' was malformed: %q", s, err)
	}
	return nil
}

// TODO Find a way to import the names of all mods here using the map from github.com/rigdev/rig/mods/allmods
// This creates a dependency cycle
var allMods = []string{
	"rigdev.annotations",
	"rigdev.datadog",
	"rigdev.env_mapping",
	"rigdev.google_cloud_sql_auth_proxy",
	"rigdev.init_container",
	"rigdev.object_template",
	"rigdev.placement",
	"rigdev.sidecar",
	"rigdev.ingress_routes",
}

func NewManager(opts ...ManagerOption) (*Manager, error) {
	manager := &Manager{
		mods: map[string]Info{},
	}

	for _, o := range opts {
		o(manager)
	}

	if manager.builtinBinaryPath == "" {
		var err error
		manager.builtinBinaryPath, err = os.Executable()
		if err != nil {
			return nil, err
		}
	}
	for _, name := range allMods {
		if err := validateModName(name); err != nil {
			return nil, err
		}
		manager.mods[name] = Info{
			IsBuiltin:  true,
			Name:       name,
			BinaryPath: manager.builtinBinaryPath,
			Args:       []string{"mod", name},
		}
	}

	p := thirdpartyModDir()
	if entries, err := os.ReadDir(p); os.IsNotExist(err) {
	} else if err != nil {
		return nil, err
	} else {
		for _, e := range entries {
			modPath := path.Join(p, e.Name())
			if err := manager.loadThirdPartyMods(modPath); err != nil {
				return nil, err
			}
		}
	}

	return manager, nil
}

var configNames = []string{"config.yaml", "config.yml", "config.json"}

func (m *Manager) loadThirdPartyMods(modPath string) error {
	entries, err := os.ReadDir(modPath)
	if os.IsNotExist(err) {
	} else if err != nil {
		return err
	}

	for _, e := range entries {
		if slices.Contains(configNames, e.Name()) {
			return m.loadThirdPartyWithConfig(modPath, entries)
		}
	}

	for _, e := range entries {
		if _, ok := m.mods[e.Name()]; ok {
			return fmt.Errorf("multiple mods with name '%s'", e.Name())
		}
		if err := validateModName(e.Name()); err != nil {
			return err
		}
		m.mods[e.Name()] = Info{
			Name:       e.Name(),
			IsBuiltin:  false,
			BinaryPath: path.Join(modPath, e.Name()),
		}
	}

	return nil
}

func (m *Manager) loadThirdPartyWithConfig(modPath string, entries []fs.DirEntry) error {
	var configPath string
	var binaryPath string

	if len(entries) != 2 {
		return fmt.Errorf("third-party mod dir contained a config file, but it did not contain exactly 2 entries")
	}

	for _, e := range entries {
		if e.IsDir() {
			return fmt.Errorf("third-party mod dir contained subdirectory %s", e.Name())
		}
		p := path.Join(modPath, e.Name())
		if slices.Contains(configNames, e.Name()) {
			if configPath != "" {
				return fmt.Errorf("third-party mod dir contained multiple config files")
			}
			configPath = p
		} else {
			binaryPath = p
		}
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var config ThirdPartyConfig
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return err
	}

	for _, mod := range config.Plugins {
		if _, ok := m.mods[mod.Name]; ok {
			return fmt.Errorf("multiple mods with name '%s'", mod.Name)
		}
		if err := validateModName(mod.Name); err != nil {
			return err
		}
		m.mods[mod.Name] = Info{
			Name:       mod.Name,
			IsBuiltin:  false,
			BinaryPath: binaryPath,
			Args:       []string{mod.Name},
		}
	}

	return nil
}

func (m *Manager) GetMod(name string) (Info, bool) {
	info, ok := m.mods[name]
	return info, ok
}

func (m *Manager) GetMods() []Info {
	mods := maps.Values(m.mods)
	slices.SortFunc(mods, func(p1, p2 Info) int {
		return strings.Compare(p1.Name, p2.Name)
	})
	return mods
}

func (m *Manager) NewStep(step v1alpha1.Step, logger logr.Logger) (*Step, error) {
	var err error
	var me []*modExecutor
	defer func() {
		if err != nil {
			for _, p := range me {
				p.Stop(context.Background())
			}
		}
	}()

	for _, mod := range step.Plugins {
		info, ok := m.mods[mod.Name]
		if !ok {
			return nil, fmt.Errorf("mod '%s' was unknown", mod.Name)
		}
		p, err := newModExecutor(
			mod.Name, step.Tag, mod.Tag, mod.Config, info.BinaryPath,
			info.Args,
			logger,
		)
		if err != nil {
			return nil, err
		}

		me = append(me, p)
	}

	matcher, err := NewMatcher(step.Namespaces, step.Capsules, step.Selector, step.EnableForPlatform)
	if err != nil {
		return nil, err
	}

	return &Step{
		step:    step,
		logger:  logger,
		mods:    me,
		matcher: matcher,
	}, nil
}
