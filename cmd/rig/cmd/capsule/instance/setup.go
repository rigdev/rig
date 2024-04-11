package instance

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	follow             bool
	tty                bool
	interactive        bool
	excludeExisting    bool
	includeDeleted     bool
	previousContainers bool
)

var since string

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
	instance := &cobra.Command{
		Use:               "instance",
		Short:             "Inspect and restart instances",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		GroupID:           capsule_cmd.TroubleshootingGroupID,
	}

	listInstances := &cobra.Command{
		Use:   "list [capsule]",
		Short: "list instances",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.list),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	listInstances.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	listInstances.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	listInstances.Flags().BoolVar(
		&includeDeleted, "include-deleted", false,
		"includes instances which have been deleted in the past 7 days",
	)
	listInstances.Flags().BoolVar(&excludeExisting, "exclude-existing", false, "only return instances which are deleted")
	instance.AddCommand(listInstances)

	getInstances := &cobra.Command{
		Use:   "get [capsule] [instance-id]",
		Short: "get instance details",
		Args:  cobra.MaximumNArgs(2),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			cli.HackCtxWrapCompletion(cmd.instanceCompletions, s),
		),
		RunE: cli.CtxWrap(cmd.get),
	}
	instance.AddCommand(getInstances)

	restartInstance := &cobra.Command{
		Use:   "restart [capsule] [instance-id]",
		Short: "Restart a single instance",
		Args:  cobra.MaximumNArgs(2),
		RunE:  cli.CtxWrap(cmd.restart),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			cli.HackCtxWrapCompletion(cmd.instanceCompletions, s),
		),
	}
	instance.AddCommand(restartInstance)

	logs := &cobra.Command{
		Use:   "logs [capsule] [instance-id]",
		Short: "Read instance logs from the capsule ",
		Args:  cobra.MaximumNArgs(2),
		RunE:  cli.CtxWrap(cmd.logs),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			cli.HackCtxWrapCompletion(cmd.instanceCompletions, s),
		),
	}
	logs.Flags().BoolVarP(
		&follow, "follow", "f", false, "keep the connection open and read out logs as they are produced",
	)
	logs.Flags().BoolVarP(
		&previousContainers, "previous-containers", "p", false,
		"Return logs from previous container terminations of the instance.",
	)
	logs.Flags().StringVarP(&since, "since", "s", "", "do not show logs older than 'since'")
	if err := logs.RegisterFlagCompletionFunc("follow", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	instance.AddCommand(logs)

	exec := &cobra.Command{
		Use:   "exec [capsule] [instance-id] -- [command] [args...]",
		Short: "Open a shell to the instance",
		RunE:  cli.CtxWrap(cmd.exec),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			cli.HackCtxWrapCompletion(cmd.instanceCompletions, s),
		),
	}
	exec.Flags().BoolVarP(&tty, "tty", "t", false, "allocate a TTY")
	exec.Flags().BoolVarP(&interactive, "interactive", "i", false, "Keep STDIN open")
	if err := exec.RegisterFlagCompletionFunc("tty", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := exec.RegisterFlagCompletionFunc("interactive", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	instance.AddCommand(exec)

	parent.AddCommand(instance)
}

func (c *Cmd) provideInstanceID(ctx context.Context, capsuleID string, arg string, argsLenAtDash int) (string, error) {
	if arg != "" && argsLenAtDash != 0 {
		return arg, nil
	}

	res, err := c.Rig.Capsule().ListInstanceStatuses(ctx, &connect.Request[capsule.ListInstanceStatusesRequest]{
		Msg: &capsule.ListInstanceStatusesRequest{
			CapsuleId:     capsuleID,
			ProjectId:     flags.GetProject(c.Scope),
			EnvironmentId: flags.GetEnvironment(c.Scope),
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

	_, s, err := c.Prompter.Select("instance", items)
	return s, err
}

func (c *Cmd) instanceCompletions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	capsule_cmd.CapsuleID = args[0]

	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	if capsule_cmd.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	var instanceIds []string

	if c.Scope.GetCurrentContext() == nil || c.Scope.GetCurrentContext().GetAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().ListInstances(ctx, &connect.Request[capsule.ListInstancesRequest]{
		Msg: &capsule.ListInstancesRequest{
			CapsuleId:     capsule_cmd.CapsuleID,
			ProjectId:     flags.GetProject(c.Scope),
			EnvironmentId: flags.GetEnvironment(c.Scope),
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
