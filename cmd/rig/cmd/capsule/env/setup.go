package env

import (
	"context"
	"strings"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	forceDeploy bool
	remove      bool
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
	env := &cobra.Command{
		Use:               "env",
		Short:             "Manage environment variables for the capsule",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	envSet := &cobra.Command{
		Use:   "set key value",
		Short: "Set an environment variable",
		Args:  cobra.ExactArgs(2),
		RunE:  base.CtxWrap(cmd.set),
	}
	envSet.Flags().BoolVarP(
		&forceDeploy,
		"force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes",
	)
	env.AddCommand(envSet)

	envGet := &cobra.Command{
		Use:   "get [key]",
		Short: "Get an environment variable",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.CtxWrap(cmd.get),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	env.AddCommand(envGet)

	envRemove := &cobra.Command{
		Use:   "remove [key]",
		Short: "Remove an environment variable",
		Args:  cobra.ExactArgs(1),
		RunE:  base.CtxWrap(cmd.remove),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	envRemove.Flags().BoolVarP(
		&forceDeploy,
		"force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes",
	)
	env.AddCommand(envRemove)

	envSource := &cobra.Command{
		Use:   "source kind name",
		Short: "Set or remove an environment source",
		Args:  cobra.ExactArgs(2),
		RunE:  base.CtxWrap(cmd.source),
	}
	envSource.Flags().BoolVarP(
		&remove,
		"remove", "r", false, "Remove the source instead of adding it",
	)
	env.AddCommand(envSource)

	parent.AddCommand(env)
}

func (c *Cmd) completions(
	ctx context.Context,
	_ *cobra.Command,
	_ []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if capsule.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	var envKeys []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	r, err := capsule.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for k := range r.GetConfig().GetContainerSettings().GetEnvironmentVariables() {
		if strings.HasPrefix(k, toComplete) {
			envKeys = append(envKeys, k)
		}
	}

	if len(envKeys) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return envKeys, cobra.ShellCompDirectiveDefault
}
