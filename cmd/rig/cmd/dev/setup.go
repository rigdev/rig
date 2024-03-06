package dev

import (
	"github.com/rigdev/rig/cmd/rig/cmd/dev/kind"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/spf13/cobra"
)

func Setup(parent *cobra.Command) {
	dev := &cobra.Command{
		Use:   "dev",
		Short: "Setup and manage development environments",
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
			auth.OmitUser:        "",
		},
	}

	kind.Setup(dev)

	parent.AddCommand(dev)
}
