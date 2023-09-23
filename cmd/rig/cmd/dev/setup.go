package dev

import (
	"github.com/rigdev/rig/cmd/rig/cmd/dev/kind"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type Cmd struct {
	fx.In

	Kind kind.Cmd
}

func (d *Cmd) Setup(parent *cobra.Command) {
	dev := &cobra.Command{
		Use: "dev",
	}

	d.Kind.Setup(dev)

	parent.AddCommand(dev)
}
