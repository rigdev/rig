package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
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
		ValidArgsFunction: common.NoCompletions,
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
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
	}
	config.AddCommand(useContext)

	parent.AddCommand(config)
}

func (c Cmd) completions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names := []string{}

	for _, ctx := range c.Cfg.Contexts {
		if strings.HasPrefix(ctx.Name, toComplete) {
			names = append(names, ctx.Name)
			names = append(names, formatContext(ctx, c.Cfg))
		}
	}

	if len(names) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return names, cobra.ShellCompDirectiveDefault
}

func formatContext(ctx *cmd_config.Context, cfg *cmd_config.Config) string {
	name := ctx.Name
	if cfg.CurrentContextName == name {
		name += "*"
	}

	return fmt.Sprintf("%v\t (Server: %v)", name, ctx.GetService().Server)
}
