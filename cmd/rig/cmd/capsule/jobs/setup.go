package jobs

import (
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
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
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	jobsGet := &cobra.Command{
		Use:   "get",
		Short: "Get cronjobs defined for the capsule",
		RunE:  base.CtxWrap(cmd.get),
	}
	jobs.AddCommand(jobsGet)

	jobsAdd := &cobra.Command{
		Use:   "add",
		Short: "Add a cronjob to the capsule",
		RunE:  base.CtxWrap(cmd.add),
	}
	jobsAdd.Flags().StringVarP(&path, "path", "p", "", "Path to a json or yaml file containing a cronjob specification")
	jobs.AddCommand(jobsAdd)

	jobsDelete := &cobra.Command{
		Use:   "delete [job-name]",
		Short: "Delete one or more cronjobs to the capsule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.CtxWrap(cmd.delete),
	}
	jobs.AddCommand(jobsDelete)

	executions := &cobra.Command{
		Use:   "executions",
		Short: "See executions of jobs",
		RunE:  base.CtxWrap(cmd.executions),
	}
	executions.Flags().StringVarP(&jobName, "job", "j", "", "Name of the job to fetch executions from")
	executions.Flags().StringVarP(&fromStr, "from", "f", "", "If set, only include executions started after this date. Layout is 2006-01-02 15:04:05")
	executions.Flags().StringVarP(&toStr, "to", "t", "", "If set, only include executions started before this date. Layout is 2006-01-02 15:04:05")
	executions.Flags().StringVarP(&since, "since", "s", "", "A duration. If set, only include executions younger than 'since'. Cannot be used if either --from or --to is used.")
	executions.Flags().StringVar(
		&states, "states", "",
		`If set, filter executions based on state. Can be a , seperated list of states.
Possible states are ongoing, completed, failed, terminated.`,
	)
	executions.Flags().Uint32VarP(
		&limit, "limit", "l", 0, "limits the number of outputs",
	)

	jobs.AddCommand(executions)

	parent.AddCommand(jobs)
}
