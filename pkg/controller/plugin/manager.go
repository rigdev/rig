package plugin

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
	"k8s.io/client-go/rest"
)

type Manager struct {
	restConfig        *rest.Config
	plugins           map[string]Info
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

func thirdpartyPluginDir() string {
	if dir, ok := os.LookupEnv("RIG_THIRDPARTY_PLUGIN_DIR"); ok {
		return dir
	}
	return "/app/bin/plugins-thirdparty"
}

func validatePluginName(s string) error {
	i := strings.Index(s, ".")
	if i == -1 {
		return fmt.Errorf(
			"plugin name '%s' was malformed. Must be of the form <group>.<name> where <group> and <name> are system names",
			s,
		)
	}

	group, name := s[:i], s[i+1:]
	if err := common.ValidateKubernetesName(group); err != nil {
		return fmt.Errorf("plugin name '%s' was malformed: %q", s, err)
	}
	if err := common.ValidateKubernetesName(name); err != nil {
		return fmt.Errorf("plugin name '%s' was malformed: %q", s, err)
	}
	return nil
}

// TODO Find a way to import the names of all plugins here using the map from  github.com/rigdev/rig/plugins/allplugins
// This creates a dependency cycle
var allPlugins = []string{
	"rigdev.service_account",
	"rigdev.deployment",
	"rigdev.ingress_routes",
	"rigdev.cron_jobs",
	"rigdev.vpa",
	"rigdev.service_monitor",

	"rigdev.annotations",
	"rigdev.datadog",
	"rigdev.env_mapping",
	"rigdev.google_cloud_sql_auth_proxy",
	"rigdev.init_container",
	"rigdev.object_template",
	"rigdev.placement",
	"rigdev.sidecar",
}

func NewManager(restCfg *rest.Config, opts ...ManagerOption) (*Manager, error) {
	manager := &Manager{
		restConfig: restCfg,
		plugins:    map[string]Info{},
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
	for _, name := range allPlugins {
		if err := validatePluginName(name); err != nil {
			return nil, err
		}
		manager.plugins[name] = Info{
			IsBuiltin:  true,
			Name:       name,
			BinaryPath: manager.builtinBinaryPath,
			Args:       []string{"plugin", name},
		}
	}

	p := thirdpartyPluginDir()
	if entries, err := os.ReadDir(p); os.IsNotExist(err) {
	} else if err != nil {
		return nil, err
	} else {
		for _, e := range entries {
			pluginPath := path.Join(p, e.Name())
			if err := manager.loadThirdPartyPlugins(pluginPath); err != nil {
				return nil, err
			}
		}
	}

	return manager, nil
}

var configNames = []string{"config.yaml", "config.yml", "config.json"}

func (m *Manager) loadThirdPartyPlugins(pluginPath string) error {
	entries, err := os.ReadDir(pluginPath)
	if os.IsNotExist(err) {
	} else if err != nil {
		return err
	}

	for _, e := range entries {
		if slices.Contains(configNames, e.Name()) {
			return m.loadThirdPartyWithConfig(pluginPath, entries)
		}
	}

	for _, e := range entries {
		if _, ok := m.plugins[e.Name()]; ok {
			return fmt.Errorf("multiple plugins with name '%s'", e.Name())
		}
		if err := validatePluginName(e.Name()); err != nil {
			return err
		}
		m.plugins[e.Name()] = Info{
			Name:       e.Name(),
			IsBuiltin:  false,
			BinaryPath: path.Join(pluginPath, e.Name()),
		}
	}

	return nil
}

func (m *Manager) loadThirdPartyWithConfig(pluginPath string, entries []fs.DirEntry) error {
	var configPath string
	var binaryPath string

	if len(entries) != 2 {
		return fmt.Errorf("third-party plugin dir contained a config file, but it did not contain exactly 2 entries")
	}

	for _, e := range entries {
		if e.IsDir() {
			return fmt.Errorf("third-party plugin dir contained subdirectory %s", e.Name())
		}
		p := path.Join(pluginPath, e.Name())
		if slices.Contains(configNames, e.Name()) {
			if configPath != "" {
				return fmt.Errorf("third-party plugin dir contained multiple config files")
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

	for _, plugin := range config.Plugins {
		if _, ok := m.plugins[plugin.Name]; ok {
			return fmt.Errorf("multiple plugins with name '%s'", plugin.Name)
		}
		if err := validatePluginName(plugin.Name); err != nil {
			return err
		}
		m.plugins[plugin.Name] = Info{
			Name:       plugin.Name,
			IsBuiltin:  false,
			BinaryPath: binaryPath,
			Args:       []string{plugin.Name},
		}
	}

	return nil
}

func (m *Manager) GetPlugin(name string) (Info, bool) {
	info, ok := m.plugins[name]
	return info, ok
}

func (m *Manager) GetPlugins() []Info {
	plugins := maps.Values(m.plugins)
	slices.SortFunc(plugins, func(p1, p2 Info) int {
		return strings.Compare(p1.Name, p2.Name)
	})
	return plugins
}

func (m *Manager) NewStep(execCtx ExecutionContext, step v1alpha1.Step, logger logr.Logger) (*Step, error) {
	var err error
	var ps []*pluginExecutor
	defer func() {
		if err != nil {
			for _, p := range ps {
				p.Stop(context.Background())
			}
		}
	}()

	for _, plugin := range step.Plugins {
		info, ok := m.plugins[plugin.Name]
		if !ok {
			return nil, fmt.Errorf("plugin '%s' was unknown", plugin.Name)
		}
		p, err := newPluginExecutor(
			execCtx,
			plugin.Name, step.Tag, plugin.Tag, plugin.Config, info.BinaryPath,
			info.Args,
			logger,
			m.restConfig,
		)
		if err != nil {
			return nil, err
		}

		ps = append(ps, p)
	}

	matcher, err := NewMatcher(MatchFromStep(step))
	if err != nil {
		return nil, err
	}

	return &Step{
		step:    step,
		logger:  logger,
		plugins: ps,
		matcher: matcher,
	}, nil
}

func MatchFromStep(step v1alpha1.Step) v1alpha1.CapsuleMatch {
	match := step.Match
	match.Namespaces = append(match.Namespaces, step.Namespaces...)
	match.Names = append(match.Names, step.Capsules...)
	match.EnableForPlatform = match.EnableForPlatform || step.EnableForPlatform
	return match
}
