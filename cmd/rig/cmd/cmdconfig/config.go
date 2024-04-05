package cmdconfig

import (
	"os"
	"path"

	"github.com/mitchellh/mapstructure"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Auth struct {
	UserID       string `yaml:"user_id,omitempty"`
	AccessToken  string `yaml:"access_token,omitempty"`
	RefreshToken string `yaml:"refresh_token,omitempty"`
}

type Context struct {
	Name          string `yaml:"name"`
	ServiceName   string `yaml:"service"`
	ProjectID     string `yaml:"project_id"`
	EnvironmentID string `yaml:"environment_id"`

	service *Service
	auth    *Auth
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

type Service struct {
	Name   string `yaml:"name,omitempty"`
	Server string `yaml:"server,omitempty"`
}

type User struct {
	Name string `yaml:"name,omitempty"`
	Auth *Auth  `yaml:"auth,omitempty"`
}

type Config struct {
	Contexts []*Context `yaml:"contexts"`

	Services []*Service `yaml:"services"`

	Users []*User `yaml:"users"`

	CurrentContextName string `yaml:"current_context"`

	filePath string
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

	if err := os.WriteFile(cfg.filePath, bs, 0o600); err != nil {
		return err
	}

	return nil
}

func NewConfig(cfgPath string) (*Config, error) {
	cfg := &Config{
		filePath: cfgPath,
	}

	if cfg.filePath != "" {
		viper.AddConfigPath(cfg.filePath)
	} else {
		configPath := os.Getenv("XDG_CONFIG_HOME")
		if configPath == "" {
			p, err := os.UserConfigDir()
			if err != nil {
				return nil, err
			}
			configPath = p
		}

		cfg.filePath = path.Join(configPath, "rig", "config.yaml")

		if _, err := os.Stat(cfg.filePath); os.IsNotExist(err) {
			if err := os.MkdirAll(path.Dir(cfg.filePath), 0o775); err != nil {
				return nil, err
			}

			if err := cfg.Save(); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
	}

	_, err := os.ReadFile(cfg.filePath)
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(cfg.filePath)

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
