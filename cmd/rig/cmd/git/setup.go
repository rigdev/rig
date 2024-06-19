package git

import (
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Prompter common.Prompter
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Prompter = c.Prompter
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	git := &cobra.Command{
		Use:   "git",
		Short: "Manage git backing of capsules",
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
			auth.OmitUser:        "",
		},
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		GroupID:           common.ManagementGroupID,
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "Get the status of the git backing",
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
			auth.OmitUser:        "",
		},
		RunE: cli.CtxWrap(cmd.status),
	}
	git.AddCommand(status)

	parent.AddCommand(git)
}
