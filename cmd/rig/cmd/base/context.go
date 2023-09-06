package base

import (
	"fmt"

	"github.com/rigdev/rig/cmd/common"
)

func UseContext(cfg *Config, name string) error {
	for _, c := range cfg.Contexts {
		if c.Name == name {
			cfg.CurrentContext = c.Name
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
		if c.Name == cfg.CurrentContext {
			labels = append(labels, c.Name+"*")
		} else {
			labels = append(labels, c.Name)
		}
	}

	n, _, err := common.PromptSelect("Context:", labels, false)
	if err != nil {
		return err
	}

	cfg.CurrentContext = cfg.Contexts[n].Name
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
		Name:    name,
		Service: name,
		User:    name,
	})

	cfg.Services = append(cfg.Services, &Service{
		Name:   name,
		Server: server,
	})

	cfg.Users = append(cfg.Users, &User{
		Name: name,
		Auth: &Auth{},
	})

	if ok, err := common.PromptConfirm("Do you want activate this context now", true); err != nil {
		return err
	} else if ok {
		cfg.CurrentContext = name
	}

	return cfg.Save()
}
