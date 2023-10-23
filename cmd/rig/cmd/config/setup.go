package config

import (
	"context"
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
			var isCurrent string
			if ctx.Name == c.Cfg.CurrentContextName {
				isCurrent = "*"
			}
			names = append(names, ctx.Name+isCurrent)
		}
	}

	if len(names) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
