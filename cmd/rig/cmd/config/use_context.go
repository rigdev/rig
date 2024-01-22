package config

import (
	"github.com/spf13/cobra"
)

func (c *Cmd) useContext(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		return c.Cfg.UseContext(args[0])
	}
	return c.Cfg.SelectContext()
}
