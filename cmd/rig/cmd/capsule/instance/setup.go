package instance

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	offset int
	limit  int
)

var (
	outputJSON bool
	follow     bool
)

func Setup(parent *cobra.Command) *cobra.Command {
	instance := &cobra.Command{
		Use:   "instance",
		Short: "Inspect and restart instances",
	}

	GetInstances := &cobra.Command{
		Use:               "get [instance-id]",
		Short:             "Get one or more instances",
		Args:              cobra.MaximumNArgs(1),
		RunE:              base.Register(get),
		ValidArgsFunction: common.Complete(cmd_capsule.InstanceCompletions, common.MaxArgsCompletionFilter(1)),
	}
	GetInstances.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	GetInstances.Flags().IntVarP(&offset, "offset", "o", 0, "offset for pagination")
	GetInstances.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	GetInstances.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	GetInstances.RegisterFlagCompletionFunc("offset", common.NoCompletions)
	GetInstances.RegisterFlagCompletionFunc("limit", common.NoCompletions)
	instance.AddCommand(GetInstances)

	restartInstance := &cobra.Command{
		Use:               "restart [instance-id]",
		Short:             "Restart a single instance",
		Args:              cobra.MaximumNArgs(1),
		RunE:              base.Register(restart),
		ValidArgsFunction: common.Complete(cmd_capsule.InstanceCompletions, common.MaxArgsCompletionFilter(1)),
	}
	instance.AddCommand(restartInstance)

	logs := &cobra.Command{
		Use:               "logs [instance-id]",
		Short:             "Read instance logs from the capsule ",
		Args:              cobra.MaximumNArgs(1),
		RunE:              base.Register(logs),
		ValidArgsFunction: common.Complete(cmd_capsule.InstanceCompletions, common.MaxArgsCompletionFilter(1)),
	}
	logs.Flags().BoolVarP(&follow, "follow", "f", false, "keep the connection open and read out logs as they are produced")
	logs.RegisterFlagCompletionFunc("follow", common.BoolCompletions)
	instance.AddCommand(logs)

	parent.AddCommand(instance)

	return instance
}

func provideInstanceID(ctx context.Context, nc rig.Client, capsuleID string, arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}

	res, err := nc.Capsule().ListInstances(ctx, &connect.Request[capsule.ListInstancesRequest]{
		Msg: &capsule.ListInstancesRequest{
			CapsuleId: capsuleID,
		},
	})
	if err != nil {
		return "", err
	}

	var items []string
	for _, i := range res.Msg.GetInstances() {
		items = append(items, i.GetInstanceId())
	}

	if len(items) == 0 {
		return "", errors.InvalidArgumentErrorf("no instances selected")
	}

	if len(items) == 1 {
		return items[0], nil
	}

	_, s, err := common.PromptSelect("instance", items)
	return s, err
}
