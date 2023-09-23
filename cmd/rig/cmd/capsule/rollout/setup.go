package rollout

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	outputJSON bool
)

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
	Cfg *cmd_config.Config
}

func (c Cmd) Setup(parent *cobra.Command) {
	rollout := &cobra.Command{
		Use:   "rollout",
		Short: "Inspect the rollouts of the capsule",
	}

	rolloutGet := &cobra.Command{
		Use:               "get [rollout-id]",
		Short:             "Get one or more rollouts",
		Args:              cobra.MaximumNArgs(1),
		RunE:              c.get,
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
	}
	rolloutGet.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	rolloutGet.Flags().IntVarP(&offset, "offset", "o", 0, "offset for pagination")
	rolloutGet.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	rolloutGet.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	rolloutGet.RegisterFlagCompletionFunc("offset", common.NoCompletions)
	rolloutGet.RegisterFlagCompletionFunc("limit", common.NoCompletions)
	rollout.AddCommand(rolloutGet)

	events := &cobra.Command{
		Use:               "events [rollout-id]",
		Short:             "List events related to a rollout, default to the current rollout",
		Args:              cobra.MaximumNArgs(1),
		RunE:              c.capsuleEvents,
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
	}
	rollout.AddCommand(events)

	parent.AddCommand(rollout)
}

func (c Cmd) completions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if capsule.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	var rolloutIds []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().ListRollouts(c.Ctx, &connect.Request[capsule_api.ListRolloutsRequest]{
		Msg: &capsule_api.ListRolloutsRequest{
			CapsuleId: capsule.CapsuleID,
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
