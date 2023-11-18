package cluster

import (
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type Cmd struct {
	fx.In

	Rig rig.Client
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
}

func Setup(parent *cobra.Command) {
	cluster := &cobra.Command{
		Use:               "cluster",
		Short:             "Manage Rig clusters",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	getConfig := &cobra.Command{
		Use:               "get-config",
		Short:             "Returns the config of the Rig cluster",
		Args:              cobra.NoArgs,
		RunE:              base.CtxWrap(cmd.get),
		ValidArgsFunction: common.NoCompletions,
	}

	cluster.AddCommand(getConfig)
	parent.AddCommand(cluster)
}
