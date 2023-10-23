package dev

import (
	"github.com/rigdev/rig/cmd/rig/cmd/dev/docker"
	"github.com/rigdev/rig/cmd/rig/cmd/dev/kind"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type Cmd struct {
	fx.In

	Kind   kind.Cmd
	Docker docker.Cmd
}

func (d *Cmd) Setup(parent *cobra.Command) {
	dev := &cobra.Command{
		Use:   "dev",
		Short: "Setup and manage development environments",
	}

	d.Kind.Setup(dev)
	d.Docker.Setup(dev)

	parent.AddCommand(dev)
}
