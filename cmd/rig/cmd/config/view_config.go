package config

import "github.com/spf13/cobra"

func (c *Cmd) viewConfig(cmd *cobra.Command, _ []string) error {
	cmd.Println(c.Cfg.Format())
	return nil
}
