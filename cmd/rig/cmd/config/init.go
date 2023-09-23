package config

import (
	"fmt"

	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
)

func (c Cmd) init(cmd *cobra.Command, args []string) error {
	if ok, err := common.PromptConfirm("Do you want to configure a new context?", true); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("aborted")
	}

	return cmd_config.CreateDefaultContext(c.Cfg)
}
