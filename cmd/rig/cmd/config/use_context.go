package config

import (
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
)

func (c Cmd) useContext(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return cmd_config.UseContext(c.Cfg, args[0])
	}

	return cmd_config.SelectContext(c.Cfg)
}
