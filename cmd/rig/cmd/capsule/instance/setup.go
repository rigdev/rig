package instance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	outputJSON bool
	follow     bool
)

var (
	since string
)

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
	Cfg *cmd_config.Config
}

func (c Cmd) Setup(parent *cobra.Command) {
	instance := &cobra.Command{
		Use:   "instance",
		Short: "Inspect and restart instances",
	}

	GetInstances := &cobra.Command{
		Use:               "get [instance-id]",
		Short:             "Get one or more instances",
		Args:              cobra.MaximumNArgs(1),
		RunE:              c.get,
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
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
		RunE:              c.restart,
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
	}
	instance.AddCommand(restartInstance)

	logs := &cobra.Command{
		Use:               "logs [instance-id]",
		Short:             "Read instance logs from the capsule ",
		Args:              cobra.MaximumNArgs(1),
		RunE:              c.logs,
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
	}
	logs.Flags().BoolVarP(&follow, "follow", "f", false, "keep the connection open and read out logs as they are produced")
	logs.Flags().StringVarP(&since, "since", "s", "1s", "do not show logs older than 'since'")
	logs.RegisterFlagCompletionFunc("follow", common.BoolCompletions)
	instance.AddCommand(logs)

	parent.AddCommand(instance)
}

func (c Cmd) provideInstanceID(ctx context.Context, capsuleID string, arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}

	res, err := c.Rig.Capsule().ListInstances(ctx, &connect.Request[capsule.ListInstancesRequest]{
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

func (c Cmd) completions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if cmd_capsule.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	var instanceIds []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().ListInstances(c.Ctx, &connect.Request[capsule.ListInstancesRequest]{
		Msg: &capsule.ListInstancesRequest{
			CapsuleId: capsule_cmd.CapsuleID,
		},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, i := range resp.Msg.GetInstances() {
		if strings.HasPrefix(fmt.Sprint(i.GetInstanceId()), toComplete) {
			instanceIds = append(instanceIds, formatInstance(i))
		}
	}

	if len(instanceIds) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return instanceIds, cobra.ShellCompDirectiveDefault
}

func formatInstance(i *capsule.Instance) string {
	var startedAt string
	if i.GetStartedAt().AsTime().IsZero() {
		startedAt = "-"
	} else {
		startedAt = time.Since(i.GetStartedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (State: %v, Started At: %v)", i.GetInstanceId(), i.GetState(), startedAt)
}
