package cmdconfig

import (
	"os"
	"path"

	"github.com/mitchellh/mapstructure"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Auth struct {
	UserID       string `json:"user_id,omitempty" yaml:"user_id,omitempty"`
	AccessToken  string `json:"access_token,omitempty" yaml:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty" yaml:"refresh_token,omitempty"`
}

type Context struct {
	Name          string `json:"name" yaml:"name"`
	ServiceName   string `json:"service" yaml:"service"`
	ProjectID     string `json:"project_id" yaml:"project_id"`
	EnvironmentID string `json:"environment_id" yaml:"environment_id"`

	service *Service
	auth    *Auth

	projectIDOverride     string
	environmentIDOverride string
}

func (c *Context) GetService() *Service {
	return c.service
}

func (c *Context) SetService(s *Service) {
	c.service = s
}

func (c *Context) GetAuth() *Auth {
	return c.auth
}

func (c *Context) SetAuth(a *Auth) {
	c.auth = a
}

func (c *Context) GetProject() string {
	if c.projectIDOverride != "" {
		return c.projectIDOverride
	}

	return c.ProjectID
}

func (c *Context) SetProjectOverride(projectID string) {
	c.projectIDOverride = projectID
}

func (c *Context) GetEnvironment() string {
	if c.environmentIDOverride != "" {
		return c.environmentIDOverride
	}

	return c.EnvironmentID
}

func (c *Context) SetEnvironmentOverride(environmentID string) {
	c.environmentIDOverride = environmentID
}

type Service struct {
	Name   string `json:"name,omitempty"`
	Server string `json:"server,omitempty"`
}

type User struct {
	Name string `json:"name,omitempty"`
	Auth *Auth  `json:"auth,omitempty"`
}

type Config struct {
	Contexts []*Context `json:"contexts" yaml:"contexts"`

	Services []*Service `json:"services" yaml:"services"`

	Users []*User `json:"users" yaml:"users"`

	CurrentContextName string `json:"current_context" yaml:"current_context"`

	filePath string
	prompter common.Prompter
	fs       afero.Fs
}

func (cfg *Config) Minify() *Config {
	return &Config{
		Contexts: []*Context{cfg.GetCurrentContext()},
		Services: []*Service{cfg.GetCurrentService()},
		Users: []*User{{
			Name: cfg.CurrentContextName,
			Auth: cfg.GetCurrentAuth(),
		}},
		CurrentContextName: cfg.CurrentContextName,
	}
}

func (cfg *Config) GetEnvironment() string {
	if c := cfg.GetCurrentContext(); c != nil {
		return c.EnvironmentID
	}
	return ""
}

func (cfg *Config) GetProject() string {
	if c := cfg.GetCurrentContext(); c != nil {
		return c.ProjectID
	}
	return ""
}

func (cfg *Config) GetCurrentContext() *Context {
	for _, c := range cfg.Contexts {
		if c.Name == cfg.CurrentContextName {
			return c
		}
	}

	return nil
}

func (cfg *Config) GetContext(name string) *Context {
	for _, c := range cfg.Contexts {
		if c.Name == name {
			if service, err := cfg.GetService(c.ServiceName); err == nil {
				c.SetService(service)
			}

			if user, err := cfg.GetUser(c.Name); err == nil {
				c.SetAuth(user.Auth)
			}

			return c
		}
	}

	return nil
}

func (cfg *Config) GetCurrentAuth() *Auth {
	c := cfg.GetCurrentContext()
	if c == nil {
		return nil
	}

	for _, u := range cfg.Users {
		if u.Name == c.Name {
			return u.Auth
		}
	}

	return nil
}

func (cfg *Config) GetUser(name string) (*User, error) {
	for _, u := range cfg.Users {
		if u.Name == name {
			return u, nil
		}
	}
	return nil, errors.NotFoundErrorf("user %s not found", name)
}

func (cfg *Config) GetCurrentService() *Service {
	c := cfg.GetCurrentContext()
	if c == nil {
		return nil
	}

	for _, cl := range cfg.Services {
		if cl.Name == c.ServiceName {
			return cl
		}
	}
	return nil
}

func (cfg *Config) GetService(name string) (*Service, error) {
	for _, cl := range cfg.Services {
		if cl.Name == name {
			return cl, nil
		}
	}
	return nil, errors.NotFoundErrorf("service %s not found", name)
}

func (cfg *Config) DeleteContext(name string) bool {
	found := false
	for idx, c := range cfg.Contexts {
		if c.Name == name {
			cfg.Contexts = append(cfg.Contexts[:idx], cfg.Contexts[idx+1:]...)
			found = true
		}
	}

	for idx, s := range cfg.Services {
		if s.Name == name {
			cfg.Services = append(cfg.Services[:idx], cfg.Services[idx+1:]...)
		}
	}

	for idx, u := range cfg.Users {
		if u.Name == name {
			cfg.Users = append(cfg.Users[:idx], cfg.Users[idx+1:]...)
		}
	}

	if found && name == cfg.CurrentContextName {
		cfg.CurrentContextName = ""
	}

	return found
}

func (cfg Config) Save() error {
	bs, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	tmpFile, err := afero.TempFile(cfg.fs, path.Dir(cfg.filePath), path.Base(cfg.filePath))
	if err != nil {
		return err
	}

	tmpName := tmpFile.Name()

	defer func() {
		_ = tmpFile.Close()
		_ = cfg.fs.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(bs); err != nil {
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	return cfg.fs.Rename(tmpName, cfg.filePath)
}

func NewConfig(cfgPath string, fs afero.Fs, p common.Prompter) (*Config, error) {
	cfg := &Config{
		filePath: cfgPath,
		prompter: p,
		fs:       fs,
	}

	if cfg.filePath != "" {
		viper.AddConfigPath(cfg.filePath)
	} else {
		var err error
		cfg.filePath, err = getConfigPath()
		if err != nil {
			return nil, err
		}
		if _, err := fs.Stat(cfg.filePath); os.IsNotExist(err) {
			if err := fs.MkdirAll(path.Dir(cfg.filePath), 0o775); err != nil {
				return nil, err
			}

			if err := cfg.Save(); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
	}

	if _, err := fs.Open(cfg.filePath); err != nil {
		return nil, err
	}
	viper.SetConfigFile(cfg.filePath)
	viper.SetFs(fs)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&cfg, viper.DecodeHook(
		uuid.MapstructureDecodeFunc(),
	), func(cfg *mapstructure.DecoderConfig) {
		cfg.TagName = "yaml"
	}); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getConfigPath() (string, error) {
	configPath := os.Getenv("XDG_CONFIG_HOME")
	if configPath == "" {
		p, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		configPath = p
	}

	return path.Join(configPath, "rig", "config.yaml"), nil
}

func NewEmptyConfig(fs afero.Fs, p common.Prompter) (*Config, error) {
	filePath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	return &Config{
		Contexts: []*Context{},
		Services: []*Service{},
		Users:    []*User{},
		prompter: p,
		fs:       fs,
		filePath: filePath,
	}, nil
}
