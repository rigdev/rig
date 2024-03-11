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
	contextName, err := PromptForContext(cfg)
	if err != nil {
		return err
	}

	cfg.CurrentContextName = contextName
	return cfg.Save()
}

func PromptForContext(cfg *Config) (string, error) {
	var labels []string
	for _, c := range cfg.Contexts {
		if c.Name == cfg.CurrentContextName {
			labels = append(labels, c.Name+"  (Current)")
		} else {
			labels = append(labels, c.Name)
		}
	}

	n, _, err := common.PromptSelect("Rig context:", labels)
	if err != nil {
		return "", err
	}

	return cfg.Contexts[n].Name, nil
}

func (cfg *Config) CreateDefaultContext() error {
	return cfg.CreateContext("local", "http://localhost:4747/")
}

func (cfg *Config) CreateContext(name, url string) error {
	var names []string
	for _, c := range cfg.Contexts {
		names = append(names, c.Name)
	}

	name, err := common.PromptInput("Name:",
		common.ValidateSystemNameOpt,
		common.InputDefaultOpt(name),
		common.ValidateUniqueOpt(names),
	)
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

	cfg.Contexts = append(cfg.Contexts, &Context{
		Name:          name,
		ServiceName:   name,
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
			UserID: uuid.Nil.String(),
		},
	})

	if ok, err := common.PromptConfirm("Do you want activate this Rig context now?", true); err != nil {
		return err
	} else if ok {
		cfg.CurrentContextName = name
	}

	return cfg.Save()
}

func (cfg *Config) CreateContextNoPrompt(name, url string) error {
	for _, c := range cfg.Contexts {
		if c.Name == name {
			return fmt.Errorf("context '%v' already exists", name)
		}
	}

	cfg.Contexts = append(cfg.Contexts, &Context{
		Name:          name,
		ServiceName:   name,
		ProjectID:     "",
		EnvironmentID: "",
	})

	cfg.Services = append(cfg.Services, &Service{
		Name:   name,
		Server: url,
	})

	cfg.Users = append(cfg.Users, &User{
		Name: name,
		Auth: &Auth{
			UserID: uuid.Nil.String(),
		},
	})

	cfg.CurrentContextName = name

	return nil
}
