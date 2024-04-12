package root

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/deploy"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/image"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/instance"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/jobs"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/rollout"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/scale"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	deploymentGroupTitle      = "Deployment Commands"
	troubleshootingGroupTitle = "Troubleshooting Commands"
	basicGroupTitle           = "Basic Commands"
)

var (
	offset int
	limit  int
)

var (
	follow             bool
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
	capsuleCmd := &cobra.Command{
		Use:   "capsule",
		Short: "Manage capsules",
		PersistentPreRunE: s.MakeInvokePreRunE(
			initCmd,
			func(ctx context.Context, cmd Cmd, c *cobra.Command, args []string) error {
				return cmd.persistentPreRunE(ctx, c, args)
			},
		),
		GroupID: common.CapsuleGroupID,
	}

	capsuleCmd.AddGroup(
		&cobra.Group{
			ID:    capsule.BasicGroupID,
			Title: basicGroupTitle,
		},
		&cobra.Group{
			ID:    capsule.DeploymentGroupID,
			Title: deploymentGroupTitle,
		},
		&cobra.Group{
			ID:    capsule.TroubleshootingGroupID,
			Title: troubleshootingGroupTitle,
		})

	capsuleCreate := &cobra.Command{
		Use:   "create [capsule]",
		Short: "Create a new capsule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.create),
		Annotations: map[string]string{
			auth.OmitCapsule:     "",
			auth.OmitEnvironment: "",
		},
		GroupID: capsule.BasicGroupID,
	}
	capsuleCmd.AddCommand(capsuleCreate)

	capsuleStop := &cobra.Command{
		Use:   "stop [capsule]",
		Short: "Stop the current rollout. This will remove all the resources related to this rollout.",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE:    cli.CtxWrap(cmd.stop),
		GroupID: capsule.DeploymentGroupID,
	}
	capsuleCmd.AddCommand(capsuleStop)

	capsuleDelete := &cobra.Command{
		Use:   "delete [capsule]",
		Short: "Delete a capsule",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
		},
		GroupID: capsule.BasicGroupID,
		RunE:    cli.CtxWrap(cmd.delete),
	}
	capsuleCmd.AddCommand(capsuleDelete)

	capsuleGet := &cobra.Command{
		Use:   "get [capsule]",
		Short: "Get a capsule",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.get),
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
		},
		GroupID: capsule.BasicGroupID,
	}
	capsuleCmd.AddCommand(capsuleGet)

	capsuleList := &cobra.Command{
		Use:   "list",
		Short: "List capsules",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.list),
		Annotations: map[string]string{
			auth.OmitCapsule: "",
		},
		GroupID: capsule.BasicGroupID,
	}
	capsuleList.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	capsuleList.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	capsuleCmd.AddCommand(capsuleList)

	capsuleLogs := &cobra.Command{
		Use:   "logs [capsule]",
		Short: "Get logs across all instances of the capsule",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE:    cli.CtxWrap(cmd.logs),
		GroupID: capsule.TroubleshootingGroupID,
	}
	capsuleLogs.Flags().BoolVarP(
		&follow, "follow", "f", false, "keep the connection open and read out logs as they are produced",
	)
	capsuleLogs.Flags().BoolVarP(
		&previousContainers, "previous-containers", "p", false,
		"Return logs from previous container terminations of the instance.",
	)
	capsuleLogs.Flags().StringVarP(&since, "since", "s", "", "do not show logs older than 'since'")
	capsuleCmd.AddCommand(capsuleLogs)

	parent.AddCommand(capsuleCmd)

	scale.Setup(capsuleCmd, s)
	image.Setup(capsuleCmd, s)
	deploy.Setup(capsuleCmd, s)
	instance.Setup(capsuleCmd, s)
	rollout.Setup(capsuleCmd, s)
	jobs.Setup(capsuleCmd, s)
}

func (c *Cmd) completions(
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

func (c *Cmd) persistentPreRunE(ctx context.Context, cmd *cobra.Command, args []string) error {
	if _, ok := cmd.Annotations[auth.OmitCapsule]; ok {
		return nil
	}

	if len(args) > 0 {
		capsule.CapsuleID = args[0]
		return nil
	}

	name, err := capsule.SelectCapsule(ctx, c.Rig, c.Prompter, c.Scope)
	if err != nil {
		return err
	}

	capsule.CapsuleID = name
	return nil
}
