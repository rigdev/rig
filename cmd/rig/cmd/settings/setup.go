package settings

import (
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Scope    scope.Scope
	Auth     *auth.Service
	Prompter common.Prompter
}

var cmd Cmd

var gitFlags common.GitFlags

func initCmd(c Cmd) {
	cmd = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	settings := &cobra.Command{
		Use:               "settings",
		Short:             "Manage Rig settings",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
			auth.OmitProject:     "",
		},
		GroupID: common.ManagementGroupID,
	}

	configuration := &cobra.Command{
		Use:   "configuration",
		Short: "View the Rig configuration",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.configuration),
	}
	settings.AddCommand(configuration)

	get := &cobra.Command{
		Use:   "get",
		Short: "Get the platform settings",
		RunE:  cli.CtxWrap(cmd.get),
		Args:  cobra.NoArgs,
	}
	settings.AddCommand(get)

	update := &cobra.Command{
		Use:   "update",
		Short: "Update the platform settings",
		RunE:  cli.CtxWrap(cmd.update),
		Args:  cobra.NoArgs,
	}
	settings.AddCommand(update)

	updateGit := &cobra.Command{
		Use:   "git",
		Short: "Update git settings",
		RunE:  cli.CtxWrap(cmd.updateGit),
		Args:  cobra.NoArgs,
	}
	gitFlags.AddFlags(updateGit)
	update.AddCommand(updateGit)

	parent.AddCommand(settings)
}
