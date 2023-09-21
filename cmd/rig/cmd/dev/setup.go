package dev

import (
	"github.com/rigdev/rig/cmd/rig/cmd/dev/kind"
	"github.com/spf13/cobra"
)

func Setup(parent *cobra.Command) {
	dev := &cobra.Command{
		Use: "dev",
	}
	kind.Setup(dev)

	parent.AddCommand(dev)
}
