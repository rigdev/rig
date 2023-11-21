package config

import (
	"fmt"

	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/spf13/cobra"
)

func (c *Cmd) init(_ *cobra.Command, _ []string) error {
	if ok, err := common.PromptConfirm("Do you want to configure a new context?", true); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("aborted")
	}

	return cmdconfig.CreateDefaultContext(c.Cfg)
}
