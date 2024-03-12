package jobs

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
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

	Rig rig.Client
	Cfg *cmdconfig.Config
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Cfg = c.Cfg
}

func Setup(parent *cobra.Command) {
	jobs := &cobra.Command{
		Use:               "jobs",
		Short:             "Manage jobs for the capsule",
		PersistentPreRunE: cli.MakeInvokePreRunE(initCmd),
	}

	jobsGet := &cobra.Command{
		Use:   "get",
		Short: "Get cronjobs defined for the capsule",
		RunE:  cli.CtxWrap(cmd.get),
	}
	jobs.AddCommand(jobsGet)

	jobsAdd := &cobra.Command{
		Use:   "add",
		Short: "Add a cronjob to the capsule",
		RunE:  cli.CtxWrap(cmd.add),
	}
	jobsAdd.Flags().StringVarP(&path, "path", "p", "", "Path to a json or yaml file containing a cronjob specification")
	jobs.AddCommand(jobsAdd)

	jobsDelete := &cobra.Command{
		Use:   "delete [job-name]",
		Short: "Delete one or more cronjobs to the capsule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.delete),
		ValidArgsFunction: common.Complete(
			cli.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	jobs.AddCommand(jobsDelete)

	executions := &cobra.Command{
		Use:   "executions",
		Short: "See executions of jobs",
		RunE:  cli.CtxWrap(cmd.executions),
	}
	executions.Flags().StringVarP(&jobName, "job", "j", "", "Name of the job to fetch executions from")
	if err := executions.RegisterFlagCompletionFunc(
		"job",
		cli.CtxWrapCompletion(cmd.completions),
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

func (c *Cmd) completions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if capsule.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	if err := cli.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var jobnames []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	r, err := capsule.GetCurrentRollout(ctx, c.Rig, c.Cfg)
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
