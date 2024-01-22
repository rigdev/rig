package cluster

import (
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type Cmd struct {
	fx.In
	Rig rig.Client
	Cfg *cmdconfig.Config
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Cfg = c.Cfg
}

func Setup(parent *cobra.Command) {
	cluster := &cobra.Command{
		Use:               "cluster",
		Short:             "Manage Rig clusters",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitProject: "",
		},
	}

	getConfig := &cobra.Command{
		Use:   "get-config",
		Short: "Returns the config of the Rig cluster",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.get),
	}

	cluster.AddCommand(getConfig)
	parent.AddCommand(cluster)
}
