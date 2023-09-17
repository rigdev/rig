package cluster

import (
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
)

func Setup(parent *cobra.Command) {
	cluster := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Rig clusters",
	}

	getConfig := &cobra.Command{
		Use:   "get-config",
		Short: "Returns the config of the Rig cluster",
		Args:  cobra.NoArgs,
		RunE:  base.Register(GetConfig),
	}

	cluster.AddCommand(getConfig)
	parent.AddCommand(cluster)
}
