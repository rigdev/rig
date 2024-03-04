package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/spf13/afero"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type Manager struct {
	builtinPath string
	plugins     map[string]Info
	fs          afero.Fs
}

type Info struct {
	Image     string `json:"image,omitempty"`
	DirPath   string `json:"dirpath,omitempty"`
	IsBuiltin bool   `json:"isBuiltin,omitempty"`

	OriginalName string `json:"originalName,omitempty"`
	Name         string `json:"name,omitempty"`
	BinaryPath   string `json:"binaryPath,omitempty"`
}

func BuiltinPluginDir() (string, error) {
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

func NewManager(fs afero.Fs) (*Manager, error) {
	pluginDir, err := BuiltinPluginDir()
	if err != nil {
		return nil, err
	}
	fmt.Println("PLUGINDIR", pluginDir)

	manager := &Manager{
		builtinPath: pluginDir,
		plugins:     map[string]Info{},
		fs:          fs,
	}

	file, err := fs.Open(pluginDir)
	if err != nil {
		return nil, err
	}
	entries, err := file.Readdir(0)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if _, ok := manager.plugins[e.Name()]; ok {
			// Should be impossible
			return nil, fmt.Errorf("builtin plugin '%s' seen twice", e.Name())
		}
		manager.plugins[e.Name()] = Info{
			IsBuiltin:    true,
			OriginalName: e.Name(),
			Name:         e.Name(),
			BinaryPath:   path.Join(pluginDir, e.Name()),
		}
	}

	file, err = fs.Open("/etc/plugins-info/contents.yaml")
	if err != nil {
		return nil, err
	}
	buffer := &bytes.Buffer{}
	if _, err := io.Copy(buffer, file); err != nil {
		return nil, err
	}

	var infos []Info
	if err := yaml.Unmarshal(buffer.Bytes(), &infos); err != nil {
		return nil, err
	}
	for _, info := range infos {
		dirPath := path.Join("/app/bin/plugins-thirdparty", info.DirPath)
		dir, err := fs.Open(dirPath)
		if err != nil {
			return nil, err
		}
		entries, err := dir.Readdir(0)
		if err != nil {
			return nil, err
		}

		var binaryName string
		for _, e := range entries {
			if !e.IsDir() {
				binaryName = e.Name()
				break
			}
		}
		if binaryName == "" {
			return nil, fmt.Errorf("plugin with image '%s' did not create a binary", info.Image)
		}
		info.BinaryPath = path.Join(dirPath, binaryName)
		info.OriginalName = binaryName
		if info.Name == "" {
			info.Name = info.OriginalName
		}
		manager.plugins[info.Name] = info
	}

	return manager, nil
}

func (m *Manager) GetPlugin(name string) (Info, bool) {
	info, ok := m.plugins[name]
	return info, ok
}

func (m *Manager) GetPlugins() []Info {
	return maps.Values(m.plugins)
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
		info, ok := m.plugins[plugin.Name]
		if !ok {
			return nil, fmt.Errorf("plugin '%s' was unknown", plugin.Name)
		}
		p, err := NewExternalPlugin(plugin.Name, info.OriginalName, plugin.Config, info.BinaryPath, logger)
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
