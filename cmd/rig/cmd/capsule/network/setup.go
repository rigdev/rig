package network

import (
	"context"
	"strings"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	outputJSON  bool
	forceDeploy bool
)

type Cmd struct {
	fx.In

	Rig rig.Client
	Cfg *cmd_config.Config
}

func Setup(parent *cobra.Command) {
	network := &cobra.Command{
		Use:   "network",
		Short: "Configure and inspect the network of the capsule",
	}

	networkConfigure := &cobra.Command{
		Use:   "configure [network-file]",
		Short: "configure the network of the capsule. If no filepath is given it goes through an interactive configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(func(c Cmd) any { return c.configure }),
	}
	networkConfigure.Flags().BoolVarP(&forceDeploy, "force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes")
	network.AddCommand(networkConfigure)

	networkGet := &cobra.Command{
		Use:   "get [name]",
		Short: "get the entire network or a specific interface of the capsule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(func(c Cmd) any { return c.get }),
		ValidArgsFunction: common.Complete(
			base.RegisterCompletion(func(c Cmd) any { return c.completions }),
			common.MaxArgsCompletionFilter(1),
		),
	}
	networkGet.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	networkGet.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	network.AddCommand(networkGet)

	parent.AddCommand(network)
}

func (c Cmd) completions(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if capsule.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	var interfaces []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	n, err := capsule.GetCurrentNetwork(ctx, c.Rig)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, i := range n.GetInterfaces() {
		if strings.HasPrefix(i.GetName(), toComplete) {
			interfaces = append(interfaces, i.GetName())
		}
	}

	if len(interfaces) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return interfaces, cobra.ShellCompDirectiveDefault
}
