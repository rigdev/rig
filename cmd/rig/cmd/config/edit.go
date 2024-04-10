package config

import (
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *CmdNoScope) editConfig(cmd *cobra.Command, args []string) error {
	var ctxName string
	var err error
	if len(args) > 0 {
		ctxName = args[0]
	} else {
		ctxName, err = c.Cfg.PromptForContext()
		if err != nil {
			return nil
		}
	}

	var ctx *cmdconfig.Context
	for _, c := range c.Cfg.Contexts {
		if c.Name == ctxName {
			ctx = c
			break
		}
	}

	if ctx == nil {
		return errors.NotFoundErrorf("context %s not found", ctxName)
	}

	if field != "" && value != "" {
		if err := applyContextChange(c.Cfg, ctx, field, value); err != nil {
			return err
		}

		if err := c.Cfg.Save(); err != nil {
			return err
		}

		cmd.Println("Context updated")
		return nil
	}

	_, field, err := c.Prompter.Select("Field to edit:", []string{"name", "server"})
	if err != nil {
		return err
	}

	defaultValue := ""
	validateFunc := common.ValidateSystemNameOpt
	switch field {
	case "name":
		defaultValue = ctx.Name
		validateFunc = common.ValidateSystemNameOpt
	case "server":
		s, err := c.Cfg.GetService(ctx.Name)
		if err != nil {
			return err
		}
		defaultValue = s.Server
		validateFunc = common.ValidateURLOpt
	}

	value, err := c.Prompter.Input("Value:", common.InputDefaultOpt(defaultValue), validateFunc)
	if err != nil {
		return err
	}

	if err := applyContextChange(c.Cfg, ctx, field, value); err != nil {
		return err
	}

	if err := c.Cfg.Save(); err != nil {
		return err
	}

	cmd.Println("Context updated")
	return nil
}

func applyContextChange(cfg *cmdconfig.Config, ctx *cmdconfig.Context, field, value string) error {
	switch field {
	case "name":
		// check if the name already exists
		for _, c := range cfg.Contexts {
			if c.Name == value {
				return errors.AlreadyExistsErrorf("context %s already exists", value)
			}
		}

		service, err := cfg.GetService(ctx.Name)
		if err != nil {
			return err
		}
		service.Name = value

		user, err := cfg.GetUser(ctx.Name)
		if err != nil {
			return err
		}
		user.Name = value
		ctx.Name = value
		ctx.ServiceName = value
	case "server":
		// check if the server host is already configured
		for _, s := range cfg.Services {
			if s.Server == value {
				return errors.AlreadyExistsErrorf("server %s is already configured in context %s", value, s.Name)
			}
		}

		service, err := cfg.GetService(ctx.Name)
		if err != nil {
			return err
		}
		service.Server = value
	default:
		return errors.NotFoundErrorf("field %s not found", field)
	}
	return nil
}
