package cmdconfig

import (
	"fmt"

	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/uuid"
)

func (cfg *Config) UseContext(name string) error {
	for _, c := range cfg.Contexts {
		if c.Name == name {
			cfg.CurrentContextName = c.Name
			return cfg.Save()
		}
	}

	return fmt.Errorf("unknown context '%v'", name)
}

func (cfg *Config) SelectContext() error {
	var labels []string
	for _, c := range cfg.Contexts {
		if c.Name == cfg.CurrentContextName {
			labels = append(labels, c.Name+"*")
		} else {
			labels = append(labels, c.Name)
		}
	}

	n, _, err := common.PromptSelect("Rig context:", labels)
	if err != nil {
		return err
	}

	cfg.CurrentContextName = cfg.Contexts[n].Name
	return cfg.Save()
}

func (cfg *Config) CreateDefaultContext() error {
	return cfg.CreateContext("local", "http://localhost:4747/")
}

func (cfg *Config) CreateContext(name, url string) error {
	name, err := common.PromptInput("Name:", common.ValidateSystemNameOpt, common.InputDefaultOpt(name))
	if err != nil {
		return err
	}

	for _, c := range cfg.Contexts {
		if c.Name == name {
			return fmt.Errorf("context '%v' already exists", name)
		}
	}

	server, err := common.PromptInput("Server:", common.ValidateURLOpt, common.InputDefaultOpt(url))
	if err != nil {
		return err
	}

	for _, s := range cfg.Services {
		if s.Server == server {
			if ok, err := common.PromptConfirm(
				"Context with this server already exists. Do you want activate this context now?", true,
			); err != nil {
				return err
			} else if ok {
				cfg.CurrentContextName = name
			}

			return cfg.Save()
		}
	}

	cfg.Contexts = append(cfg.Contexts, &Context{
		Name:          name,
		ServiceName:   name,
		UserName:      name,
		ProjectID:     "",
		EnvironmentID: "",
	})

	cfg.Services = append(cfg.Services, &Service{
		Name:   name,
		Server: server,
	})

	cfg.Users = append(cfg.Users, &User{
		Name: name,
		Auth: &Auth{
			UserID: uuid.Nil,
		},
	})

	if ok, err := common.PromptConfirm("Do you want activate this Rig context now?", true); err != nil {
		return err
	} else if ok {
		cfg.CurrentContextName = name
	}

	return cfg.Save()
}
