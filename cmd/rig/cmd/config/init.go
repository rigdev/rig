package config

import (
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *CmdNoScope) init(_ *cobra.Command, _ []string) error {
	if c.PromptInfo.ContextCreation {
		return nil
	}

	return c.Cfg.CreateContextAndSave(contextName, flags.Flags.Host, bool(c.Interactive))
}
