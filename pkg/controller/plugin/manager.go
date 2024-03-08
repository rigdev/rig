package plugin

import (
	"context"
	"fmt"
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
	builtinPath string
	plugins     map[string]Info
}

type Info struct {
	Type       string
	Image      string
	IsBuiltin  bool
	BinaryPath string
}

func builtinPluginDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	pluginDir := path.Join(path.Dir(execPath), "plugin")
	if dir := os.Getenv("RIG_PLUGIN_DIR"); dir != "" {
		pluginDir = dir
	}
	return pluginDir, nil
}

func thirdpartyPluginDir() string {
	if dir, ok := os.LookupEnv("RIG_THIRDPARTY_PLUGIN_DIR"); ok {
		return dir
	}
	return "/app/bin/plugins-thirdparty"
}

func validatePluginType(s string) error {
	i := strings.Index(s, ".")
	if i == -1 {
		return fmt.Errorf(
			"plugin type '%s' was malformed. Must be of the form <group>.<type> where <group> and <type> are system names",
			s,
		)
	}

	group, type_ := s[:i], s[i+1:]
	if err := common.ValidateKubernetesName(group); err != nil {
		return fmt.Errorf("plugin type '%s' was malformed: %q", s, err)
	}
	if err := common.ValidateKubernetesName(type_); err != nil {
		return fmt.Errorf("plugin type '%s' was malformed: %q", s, err)
	}
	return nil
}

func NewManager() (*Manager, error) {
	pluginDir, err := builtinPluginDir()
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		builtinPath: pluginDir,
		plugins:     map[string]Info{},
	}

	if entries, err := os.ReadDir(pluginDir); os.IsNotExist(err) {
	} else if err != nil {
		return nil, err
	} else {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if _, ok := manager.plugins[e.Name()]; ok {
				// Should be impossible
				return nil, fmt.Errorf("builtin plugin '%s' seen twice", e.Name())
			}
			if err := validatePluginType(e.Name()); err != nil {
				return nil, err
			}
			manager.plugins[e.Name()] = Info{
				IsBuiltin:  true,
				Type:       e.Name(),
				BinaryPath: path.Join(pluginDir, e.Name()),
			}
		}
	}

	p := thirdpartyPluginDir()
	if entries, err := os.ReadDir(p); os.IsNotExist(err) {
	} else if err != nil {
		return nil, err
	} else {
		for _, e := range entries {
			pluginPath := path.Join(p, e.Name())
			entries, err := os.ReadDir(pluginPath)
			if err != nil {
				return nil, err
			}
			var types []string
			for _, ee := range entries {
				if ee.Name() == "manifest.yaml" {
					continue
				}
				if _, ok := manager.plugins[ee.Name()]; ok {
					return nil, fmt.Errorf("multiple plugins with type '%s'", ee.Name())
				}
				if err := validatePluginType(ee.Name()); err != nil {
					return nil, err
				}
				manager.plugins[ee.Name()] = Info{
					Type:       ee.Name(),
					IsBuiltin:  false,
					BinaryPath: path.Join(pluginPath, ee.Name()),
				}
				types = append(types, ee.Name())
			}
			bytes, err := os.ReadFile(path.Join(pluginPath, "manifest.yaml"))
			if os.IsNotExist(err) {
				continue
			} else if err != nil {
				return nil, err
			}

			manifest := struct {
				Image string `json:"image,omitempty"`
			}{}
			if err := yaml.Unmarshal(bytes, &manifest); err != nil {
				return nil, err
			}
			for _, type_ := range types {
				info := manager.plugins[type_]
				info.Image = manifest.Image
				manager.plugins[type_] = info
			}
		}
	}

	return manager, nil
}

func (m *Manager) GetPlugin(type_ string) (Info, bool) {
	info, ok := m.plugins[type_]
	return info, ok
}

func (m *Manager) GetPlugins() []Info {
	plugins := maps.Values(m.plugins)
	slices.SortFunc(plugins, func(p1, p2 Info) int {
		return strings.Compare(p1.Type, p2.Type)
	})
	return plugins
}

func (m *Manager) NewStep(step v1alpha1.Step, logger logr.Logger) (*Step, error) {
	var err error
	var ps []Plugin
	defer func() {
		if err != nil {
			for _, p := range ps {
				p.Stop(context.Background())
			}
		}
	}()

	for _, plugin := range step.Plugins {
		info, ok := m.plugins[plugin.Type]
		if !ok {
			return nil, fmt.Errorf("plugin '%s' was unknown", plugin.Type)
		}
		p, err := NewExternalPlugin(
			plugin.Type, step.Name, plugin.Name, plugin.Config, info.BinaryPath,
			logger,
		)
		if err != nil {
			return nil, err
		}

		ps = append(ps, p)
	}

	matcher, err := NewMatcher(step.Namespaces, step.Capsules, step.Selector)
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
