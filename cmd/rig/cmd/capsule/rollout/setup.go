package rollout

import (
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

var (
	offset int
	limit  int
)

var (
	outputJSON bool
)

var ()

func Setup(parent *cobra.Command) *cobra.Command {
	rollout := &cobra.Command{
		Use:   "rollout",
		Short: "Inspect the rollouts of the capsule",
	}

	rolloutGet := &cobra.Command{
		Use:               "get [rollout-id]",
		Short:             "Get one or more rollouts",
		Args:              cobra.MaximumNArgs(1),
		RunE:              base.Register(get),
		ValidArgsFunction: common.Complete(capsule.RolloutCompletions, common.MaxArgsCompletionFilter(1)),
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
		RunE:              base.Register(CapsuleEvents),
		ValidArgsFunction: common.Complete(capsule.RolloutCompletions, common.MaxArgsCompletionFilter(1)),
	}
	rollout.AddCommand(events)

	parent.AddCommand(rollout)

	return rollout
}
