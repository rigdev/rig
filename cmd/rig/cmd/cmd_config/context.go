package cmd_config

import (
	"fmt"

	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/uuid"
)

func UseContext(cfg *Config, name string) error {
	for _, c := range cfg.Contexts {
		if c.Name == name {
			cfg.CurrentContextName = c.Name
			return cfg.Save()
		}
	}

	return fmt.Errorf("unknown context '%v'", name)
}

func SelectContext(cfg *Config) error {
	var names []string
	var labels []string
	for _, c := range cfg.Contexts {
		names = append(names, c.Name)
		if c.Name == cfg.CurrentContextName {
			labels = append(labels, c.Name+"*")
		} else {
			labels = append(labels, c.Name)
		}
	}

	n, _, err := common.PromptSelect("Context:", labels, false)
	if err != nil {
		return err
	}

	cfg.CurrentContextName = cfg.Contexts[n].Name
	return cfg.Save()
}

func CreateContext(cfg *Config) error {
	name, err := common.PromptGetInputWithDefault("Name:", common.ValidateSystemName, "local")
	if err != nil {
		return err
	}

	server, err := common.PromptGetInputWithDefault("Server:", common.ValidateURL, "http://localhost:4747/")
	if err != nil {
		return err
	}

	cfg.Contexts = append(cfg.Contexts, &Context{
		Name:        name,
		ServiceName: name,
		UserName:    name,
		Project: struct {
			ProjectID    uuid.UUID `yaml:"project_id"`
			ProjectToken string    `yaml:"project_token"`
		}{
			ProjectID:    uuid.Nil,
			ProjectToken: "",
		},
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

	if ok, err := common.PromptConfirm("Do you want activate this context now", true); err != nil {
		return err
	} else if ok {
		cfg.CurrentContextName = name
	}

	return cfg.Save()
}
