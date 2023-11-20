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
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
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
	follow      bool
	tty         bool
	interactive bool
)

var (
	since string
)

type Cmd struct {
	fx.In

	Rig rig.Client
	Cfg *cmd_config.Config
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Cfg = c.Cfg
}

func Setup(parent *cobra.Command) {
	instance := &cobra.Command{
		Use:               "instance",
		Short:             "Inspect and restart instances",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	GetInstances := &cobra.Command{
		Use:   "get [instance-id]",
		Short: "Get one or more instances",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.CtxWrap(cmd.get),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	GetInstances.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	GetInstances.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	instance.AddCommand(GetInstances)

	restartInstance := &cobra.Command{
		Use:   "restart [instance-id]",
		Short: "Restart a single instance",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.CtxWrap(cmd.restart),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	instance.AddCommand(restartInstance)

	logs := &cobra.Command{
		Use:   "logs [instance-id]",
		Short: "Read instance logs from the capsule ",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.CtxWrap(cmd.logs),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	logs.Flags().BoolVarP(&follow, "follow", "f", false, "keep the connection open and read out logs as they are produced")
	logs.Flags().StringVarP(&since, "since", "s", "1s", "do not show logs older than 'since'")
	logs.RegisterFlagCompletionFunc("follow", common.BoolCompletions)
	instance.AddCommand(logs)

	exec := &cobra.Command{
		Use:   "exec [instance-id] -- [command] [args...]",
		Short: "Open a shell to the instance",
		RunE:  base.CtxWrap(cmd.exec),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	exec.Flags().BoolVarP(&tty, "tty", "t", false, "allocate a TTY")
	exec.Flags().BoolVarP(&interactive, "interactive", "i", false, "Keep STDIN open")
	exec.RegisterFlagCompletionFunc("tty", common.BoolCompletions)
	exec.RegisterFlagCompletionFunc("interactive", common.BoolCompletions)
	instance.AddCommand(exec)

	parent.AddCommand(instance)
}

func (c *Cmd) provideInstanceID(ctx context.Context, capsuleID string, arg string, argsLenAtDash int) (string, error) {
	if arg != "" && argsLenAtDash != 0 {
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

func (c *Cmd) completions(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if capsule_cmd.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	var instanceIds []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().ListInstances(ctx, &connect.Request[capsule.ListInstancesRequest]{
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
