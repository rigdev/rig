package rollout

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset    int
	limit     int
	rolloutID int
)

var (
	forceDeploy bool
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
	rollout := &cobra.Command{
		Use:               "rollout",
		Short:             "Inspect the rollouts of the capsule",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	rolloutGet := &cobra.Command{
		Use:   "get [rollout-id]",
		Short: "Get one or more rollouts",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.CtxWrap(cmd.get),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	rolloutGet.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	rolloutGet.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	rollout.AddCommand(rolloutGet)

	events := &cobra.Command{
		Use:   "events [rollout-id]",
		Short: "List events related to a rollout, default to the current rollout",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.CtxWrap(cmd.capsuleEvents),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	rollout.AddCommand(events)

	rollback := &cobra.Command{
		Use:   "rollback [rollout-id]",
		Short: "Rollback the capsule to a previous rollout",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.rollback),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	rollback.Flags().BoolVarP(
		&forceDeploy,
		"force-deploy", "f", false, "Abort the current rollout if one is in progress and perform the rollback",
	)
	rollback.Flags().IntVarP(
		&rolloutID,
		"rollout-id",
		"r", -1, "The rollout to rollback to. If not given, will roll back to the latest successful rollout.",
	)
	rollout.AddCommand(rollback)

	parent.AddCommand(rollout)
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

	var rolloutIds []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().ListRollouts(ctx, &connect.Request[capsule_api.ListRolloutsRequest]{
		Msg: &capsule_api.ListRolloutsRequest{
			CapsuleId:     capsule.CapsuleID,
			ProjectId:     c.Cfg.GetProject(),
			EnvironmentId: base.Flags.Environment,
		},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, r := range resp.Msg.GetRollouts() {
		if strings.HasPrefix(fmt.Sprint(r.GetRolloutId()), toComplete) {
			rolloutIds = append(rolloutIds, formatRollout(r))
		}
	}

	if len(rolloutIds) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return rolloutIds, cobra.ShellCompDirectiveDefault
}

func formatRollout(r *capsule_api.Rollout) string {
	return fmt.Sprintf("%v\t (State: %v)", r.GetRolloutId(), r.GetStatus().GetState())
}
