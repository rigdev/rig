package dev

import (
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/dev/kind"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/spf13/cobra"
)

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	dev := &cobra.Command{
		Use:   "dev",
		Short: "Setup and manage development environments",
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
			auth.OmitUser:        "",
		},
		GroupID: common.OtherGroupID,
	}

	kind.Setup(dev, s)

	parent.AddCommand(dev)
}
