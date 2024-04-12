package main

import (
	"fmt"

	"github.com/rigdev/rig/mods/allmods"
	"github.com/rigdev/rig/pkg/controller/mod"
	"github.com/spf13/cobra"
)

func modSetup(parent *cobra.Command) {
	modCmd := &cobra.Command{
		Use:     "mod",
		Aliases: []string{"plugin"},
		Short:   "Execute a builtin mod",
		Args:    cobra.ExactArgs(1),
		RunE:    runMod,
	}
	parent.AddCommand(modCmd)
}

func runMod(_ *cobra.Command, args []string) error {
	modName := args[0]
	p, ok := allmods.Mods[modName]
	if !ok {
		return fmt.Errorf("unknown mod name %s", modName)
	}
	mod.StartMod(modName, p)
	return nil
}
