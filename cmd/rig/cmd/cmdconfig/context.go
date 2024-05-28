package cmdconfig

import (
	"fmt"
	"net/url"

	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
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
	contextName, err := cfg.PromptForContext()
	if err != nil {
		return err
	}

	cfg.CurrentContextName = contextName
	return cfg.Save()
}

func (cfg *Config) PromptForContext() (string, error) {
	var labels []string
	for _, c := range cfg.Contexts {
		if c.Name == cfg.CurrentContextName {
			labels = append(labels, c.Name+"  (Current)")
		} else {
			labels = append(labels, c.Name)
		}
	}

	n, _, err := cfg.prompter.Select("Rig context:", labels)
	if err != nil {
		return "", err
	}

	return cfg.Contexts[n].Name, nil
}

func (cfg *Config) CreateContextAndSave(name, host string, interactive bool) error {
	if err := cfg.CreateContext(name, host, interactive); err != nil {
		return err
	}

	return cfg.Save()
}

func (cfg *Config) CreateContext(name, host string, interactive bool) error {
	var err error

	if name == "" {
		if !interactive {
			return fmt.Errorf("no context provided, use `--context` to specify")
		}

		var names []string
		for _, c := range cfg.Contexts {
			names = append(names, c.Name)
		}

		name, err = cfg.prompter.Input("Name:",
			common.ValidateSystemNameOpt,
			common.InputDefaultOpt(name),
			common.ValidateUniqueOpt(names),
		)
		if err != nil {
			return err
		}
	}

	for _, c := range cfg.Contexts {
		if c.Name == name {
			return fmt.Errorf("context '%v' already exists", name)
		}
	}

	if host == "" {
		if !interactive {
			return fmt.Errorf("no host provided, use `--host` or `RIG_HOST` to specify the host of the Rig platform")
		}

		host, err = cfg.prompter.Input("Host (Platform URL):", common.ValidateURLOpt, common.InputDefaultOpt(host))
		if err != nil {
			return err
		}
	}

	url, err := url.Parse(host)
	if err != nil {
		return errors.InvalidArgumentErrorf("invalid host, must be a fully qualified URL: %v", err)
	}

	if url.Host == "" {
		return errors.InvalidArgumentErrorf("invalid host, must be a fully qualified URL: missing hostname")
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return errors.InvalidArgumentErrorf("invalid host, must start with `https://` or `http://`")
	}

	svc := &Service{
		Name:   name,
		Server: host,
	}
	auth := &Auth{
		UserID: uuid.Nil.String(),
	}
	cfg.Contexts = append(cfg.Contexts, &Context{
		Name:          name,
		ServiceName:   name,
		ProjectID:     "",
		EnvironmentID: "",
		service:       svc,
		auth:          auth,
	})

	cfg.Services = append(cfg.Services, svc)

	cfg.Users = append(cfg.Users, &User{
		Name: name,
		Auth: auth,
	})

	if interactive {
		if ok, err := cfg.prompter.Confirm("Do you want activate this Rig context now?", true); err != nil {
			return err
		} else if ok {
			cfg.CurrentContextName = name
		}
	} else {
		if len(cfg.CurrentContextName) == 0 {
			cfg.CurrentContextName = name
		}
	}

	return nil
}
