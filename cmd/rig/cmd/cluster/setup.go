package cluster

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
	Rig   rig.Client
	Scope scope.Scope
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Scope = c.Scope
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	cluster := &cobra.Command{
		Use:               "cluster",
		Short:             "Manage Rig clusters",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitProject: "",
		},
		GroupID: common.ManagementGroupID,
	}

	getConfig := &cobra.Command{
		Use:   "get-config",
		Short: "Returns the config of the Rig cluster",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.get),
	}

	cluster.AddCommand(getConfig)
	parent.AddCommand(cluster)
}
