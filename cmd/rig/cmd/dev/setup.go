package dev

import (
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/dev/kind"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	hostInterface    []string
	capsuleInterface []string
	filePath         string
	capsuleName      string
	printConfig      bool
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Scope    scope.Scope
	Prompter common.Prompter
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd = c
}

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

	host := &cobra.Command{
		Use:               "host",
		Short:             "Configure a Capsule to tunnel traffic to/from the host machine",
		Args:              cobra.NoArgs,
		RunE:              cli.CtxWrap(cmd.host),
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
	}

	host.Flags().StringArrayVar(
		&hostInterface,
		"host-interface", nil, "Host interface to tunnel into Kubernetes.",
	)
	host.Flags().StringArrayVar(
		&capsuleInterface,
		"capsule-interface", nil, "Capsule interface to tunnel to host traffic.",
	)
	host.Flags().StringVarP(
		&filePath,
		"path", "f", "", "Path to a file containing a HostCapsule spec.",
	)
	host.Flags().StringVarP(
		&capsuleName,
		"capsule", "c", "", "Name of capsule to configure as.",
	)
	host.Flags().BoolVar(
		&printConfig,
		"print", false, "Print the HostCapsule configuration file as configured and exit.",
	)

	dev.AddCommand(host)

	kind.Setup(dev, s)

	parent.AddCommand(dev)
}
