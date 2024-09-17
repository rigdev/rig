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
	remoteContext    string
	interfaces       []string
	proxyTag         string
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
	host.Flags().StringVar(
		&proxyTag,
		"tag", "latest", "Specify the image tag of rig-proxy to use.",
	)

	dev.AddCommand(host)

	remote := &cobra.Command{
		Use:               "remote",
		Short:             "Connect a remote Capsule from different context, to the local dev environment",
		Args:              cobra.NoArgs,
		RunE:              cli.CtxWrap(cmd.remote),
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
	}

	remote.Flags().StringArrayVarP(
		&interfaces,
		"interface", "i", nil,
		"Capsules interface to tunnel. Default is to forward all. Can both given as both name and port.",
	)
	remote.Flags().StringVarP(
		&remoteContext,
		"remote-context", "r", "", "The remote context to connect to",
	)
	remote.Flags().StringVarP(
		&capsuleName,
		"capsule", "c", "", "Name of capsule to forward.",
	)
	remote.MarkFlagsOneRequired("remote-context")

	dev.AddCommand(remote)

	kind.Setup(dev, s)

	parent.AddCommand(dev)
}
