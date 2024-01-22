package config

import (
	"fmt"

	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) init(_ *cobra.Command, _ []string) error {
	if c.PromptInfo.ContextCreation {
		return nil
	}

	if ok, err := common.PromptConfirm("Do you want to configure a new context?", true); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("aborted")
	}

	return c.Cfg.CreateDefaultContext()
}
