package config

import (
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/spf13/cobra"
)

func (c *Cmd) useContext(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		return cmdconfig.UseContext(c.Cfg, args[0])
	}

	return cmdconfig.SelectContext(c.Cfg)
}
