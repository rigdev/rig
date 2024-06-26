package config

import (
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *CmdNoScope) delete(cmd *cobra.Command, args []string) error {
	var ctx string
	var err error
	if len(args) == 0 {
		ctx, err = c.Cfg.PromptForContext()
		if err != nil {
			return err
		}
	} else {
		ctx = args[0]
	}

	deleted := c.Cfg.DeleteContext(ctx)
	if !deleted {
		return errors.NotFoundErrorf("context %s not found", ctx)
	}
	err = c.Cfg.Save()
	if err != nil {
		return err
	}
	cmd.Println("Context deleted")
	return nil
}
