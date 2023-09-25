package config

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
	Cfg *cmd_config.Config
}

func (c Cmd) Setup(parent *cobra.Command) {
	config := &cobra.Command{
		Use:   "config",
		Short: "Manage Rig CLI configuration",
	}

	init := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new context",
		Args:  cobra.NoArgs,
		RunE:  c.init,
		Annotations: map[string]string{
			base.OmitProject: "",
			base.OmitUser:    "",
		},
	}
	config.AddCommand(init)

	useContext := &cobra.Command{
		Use:   "use-context [context]",
		Short: "Change the current context to use",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.useContext,
		Annotations: map[string]string{
			base.OmitProject: "",
			base.OmitUser:    "",
		},
	}
	config.AddCommand(useContext)

	parent.AddCommand(config)
}
