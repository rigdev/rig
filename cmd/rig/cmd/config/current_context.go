package config

import "github.com/spf13/cobra"

func (c *Cmd) currentContext(cmd *cobra.Command, _ []string) error {
	cmd.Println(c.Cfg.CurrentContextName)
	return nil
}
