package config

import "github.com/spf13/cobra"

func (c *CmdWScope) currentContext(cmd *cobra.Command, _ []string) error {
	cmd.Println(c.Scope.GetCurrentContext().Name)
	return nil
}
