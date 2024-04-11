package jobs

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	jobName string
	fromStr string
	toStr   string
	states  string
	since   string
	limit   uint32

	path string
)

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
	jobs := &cobra.Command{
		Use:               "jobs",
		Short:             "Manage jobs for the capsule",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		GroupID:           capsule.DeploymentGroupID,
	}

	jobsList := &cobra.Command{
		Use:               "list [capsule]",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s)),
		Short:             "List cronjobs defined for the capsule",
		RunE:              cli.CtxWrap(cmd.list),
	}
	jobs.AddCommand(jobsList)

	jobsAdd := &cobra.Command{
		Use:               "add [capsule]",
		Short:             "Add a cronjob to the capsule",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s)),
		RunE:              cli.CtxWrap(cmd.add),
	}
	jobsAdd.Flags().StringVarP(&path, "path", "p", "", "Path to a json or yaml file containing a cronjob specification")
	jobs.AddCommand(jobsAdd)

	jobsDelete := &cobra.Command{
		Use:   "delete [capsule] [job-name]",
		Short: "Delete one or more cronjobs to the capsule",
		Args:  cobra.MaximumNArgs(2),
		RunE:  cli.CtxWrap(cmd.delete),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			cli.HackCtxWrapCompletion(cmd.jobCompletions, s),
		),
	}
	jobs.AddCommand(jobsDelete)

	executions := &cobra.Command{
		Use:   "executions [capsule]",
		Short: "See executions of jobs",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.executions),
	}
	executions.Flags().StringVarP(&jobName, "job", "j", "", "Name of the job to fetch executions from")
	if err := executions.RegisterFlagCompletionFunc(
		"job",
		cli.HackCtxWrapCompletion(cmd.jobCompletions, s),
	); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	executions.Flags().StringVarP(
		&fromStr, "from", "f", "",
		"If set, only include executions started after this date. Layout is 2006-01-02 15:04:05",
	)
	executions.Flags().StringVarP(
		&toStr, "to", "t", "",
		"If set, only include executions started before this date. Layout is 2006-01-02 15:04:05",
	)
	executions.Flags().StringVarP(
		&since, "since", "s", "",
		"A duration. If set, only include executions younger than 'since'. Cannot be used if either --from or --to is used.",
	)
	executions.Flags().StringVar(
		&states, "states", "",
		`If set, filter executions based on state. Can be a , seperated list of states.
Possible states are ongoing, completed, failed, terminated.`,
	)

	if err := executions.RegisterFlagCompletionFunc("states",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"ongoing", "completed", "failed", "terminated"}, cobra.ShellCompDirectiveDefault
		}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	executions.Flags().Uint32VarP(
		&limit, "limit", "l", 50, "limits the number of outputs",
	)

	jobs.AddCommand(executions)

	parent.AddCommand(jobs)
}

func (c *Cmd) jobCompletions(
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

	var jobnames []string

	if c.Scope.GetCurrentContext() == nil || c.Scope.GetCurrentContext().GetAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	r, err := capsule.GetCurrentRollout(ctx, c.Rig, c.Scope)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, job := range r.GetConfig().GetCronJobs() {
		if strings.HasPrefix(job.GetJobName(), toComplete) {
			jobnames = append(jobnames, job.GetJobName())
		}
	}

	if len(jobnames) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return jobnames, cobra.ShellCompDirectiveDefault
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
