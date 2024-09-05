package rollout

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var follow bool

var forceDeploy bool

type Cmd struct {
	fx.In

	Rig      rig.Client
	Scope    scope.Scope
	Prompter common.Prompter
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	rollout := &cobra.Command{
		Use:               "rollout",
		Short:             "Inspect the rollouts of the capsule",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		GroupID:           capsule.TroubleshootingGroupID,
	}

	rolloutList := &cobra.Command{
		Use:   "list [capsule]",
		Short: "List rollouts",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.list),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	rolloutList.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	rolloutList.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	rolloutList.Flags().BoolVarP(&follow, "follow", "f", false,
		"keep the connection open and read out rollouts as it they are updated.")
	rolloutList.MarkFlagsMutuallyExclusive("follow", "offset")
	rollout.AddCommand(rolloutList)

	rollback := &cobra.Command{
		Use:   "rollback [capsule-id] [rollout-id]",
		Short: "Rollback the capsule to a previous rollout",
		Args:  cobra.MaximumNArgs(2),
		RunE:  cli.CtxWrap(cmd.rollback),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			cli.HackCtxWrapCompletion(cmd.rolloutCompletions, s),
		),
	}
	rollback.Flags().BoolVarP(
		&forceDeploy,
		"force-deploy", "f", false, "Abort the current rollout if one is in progress and perform the rollback",
	)
	rollout.AddCommand(rollback)

	parent.AddCommand(rollout)
}

func (c *Cmd) rolloutCompletions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	capsule.CapsuleID = args[0]

	if capsule.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var rolloutIDs []string

	if c.Scope.GetCurrentContext() == nil || c.Scope.GetCurrentContext().GetAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().ListRollouts(ctx, &connect.Request[capsule_api.ListRolloutsRequest]{
		Msg: &capsule_api.ListRolloutsRequest{
			CapsuleId:     capsule.CapsuleID,
			ProjectId:     c.Scope.GetCurrentContext().GetProject(),
			EnvironmentId: c.Scope.GetCurrentContext().GetEnvironment(),
		},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, r := range resp.Msg.GetRollouts() {
		if strings.HasPrefix(fmt.Sprint(r.GetRolloutId()), toComplete) {
			rolloutIDs = append(rolloutIDs, formatRollout(r))
		}
	}

	if len(rolloutIDs) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return rolloutIDs, cobra.ShellCompDirectiveDefault
}

func formatRollout(r *capsule_api.Rollout) string {
	return fmt.Sprintf("%v\t (State: %v)", r.GetRolloutId(), r.GetStatus().GetState())
}

func (c *Cmd) capsuleCompletions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions.Capsules(ctx, c.Rig, toComplete, c.Scope)
}
