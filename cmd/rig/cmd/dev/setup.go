package dev

import (
	"github.com/rigdev/rig/cmd/rig/cmd/dev/docker"
	"github.com/rigdev/rig/cmd/rig/cmd/dev/kind"
	"github.com/spf13/cobra"
)

func Setup(parent *cobra.Command) {
	dev := &cobra.Command{
		Use:   "dev",
		Short: "Setup and manage development environments",
	}

	kind.Setup(dev)
	docker.Setup(dev)

	parent.AddCommand(dev)
}
