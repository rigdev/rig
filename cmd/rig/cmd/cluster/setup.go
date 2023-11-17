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

func Setup(parent *cobra.Command) {
	cluster := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Rig clusters",
	}

	getConfig := &cobra.Command{
		Use:               "get-config",
		Short:             "Returns the config of the Rig cluster",
		Args:              cobra.NoArgs,
		RunE:              base.Register(func(c Cmd) any { return c.get }),
		ValidArgsFunction: common.NoCompletions,
	}

	cluster.AddCommand(getConfig)
	parent.AddCommand(cluster)
}
