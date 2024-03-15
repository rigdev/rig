package main

import (
	"fmt"

	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/plugins/allplugins"
	"github.com/spf13/cobra"
)

func pluginSetup(parent *cobra.Command) {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Execute a builtin plugin",
		Args:  cobra.ExactArgs(1),
		RunE:  runPlugin,
	}
	parent.AddCommand(pluginCmd)
}

func runPlugin(_ *cobra.Command, args []string) error {
	pluginName := args[0]
	p, ok := allplugins.Plugins[pluginName]
	if !ok {
		return fmt.Errorf("unknown plugin name %s", pluginName)
	}
	plugin.StartPlugin(pluginName, p)
	return nil
}
