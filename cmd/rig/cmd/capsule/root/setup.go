package root

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/docker/docker/client"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/deploy"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/image"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/instance"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/jobs"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/rollout"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/scale"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	interactive        bool
	forceDeploy        bool
	follow             bool
	previousContainers bool
)

var since string

type Cmd struct {
	fx.In

	Rig          rig.Client
	Scope        scope.Scope
	DockerClient *client.Client
	Prompter     common.Prompter
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
	}
	capsuleCmd.PersistentFlags().StringVarP(&capsule.CapsuleID, "capsule-id", "c", "", "Id of the capsule")
	if err := capsuleCmd.RegisterFlagCompletionFunc(
		"capsule-id",
		cli.HackCtxWrapCompletion(cmd.completions, s),
	); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	capsuleCreate := &cobra.Command{
		Use:   "create",
		Short: "Create a new capsule",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.create),
		Annotations: map[string]string{
			auth.OmitCapsule: "",
		},
	}
	capsuleCreate.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactive mode")
	capsuleCreate.Flags().BoolVarP(
		&forceDeploy,
		"force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes",
	)
	if err := capsuleCreate.RegisterFlagCompletionFunc("interactive", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := capsuleCreate.RegisterFlagCompletionFunc("force-deploy", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	capsuleCmd.AddCommand(capsuleCreate)

	capsuleAbort := &cobra.Command{
		Use:   "abort",
		Short: "Abort the current rollout. This will leave the capsule in a undefined state",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.abort),
	}
	capsuleCmd.AddCommand(capsuleAbort)

	capsuleDelete := &cobra.Command{
		Use:   "delete",
		Short: "Delete a capsule",
		Args:  cobra.NoArgs,
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
		},
		RunE: cli.CtxWrap(cmd.delete),
	}
	capsuleCmd.AddCommand(capsuleDelete)

	capsuleGet := &cobra.Command{
		Use:               "get",
		Short:             "Get one or more capsules",
		PersistentPreRunE: s.PersistentPreRunE,
		Args:              cobra.NoArgs,
		Annotations: map[string]string{
			auth.OmitCapsule: "",
		},
		RunE: cli.CtxWrap(cmd.get),
	}
	capsuleGet.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	capsuleGet.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	capsuleCmd.AddCommand(capsuleGet)

	capsuleLogs := &cobra.Command{
		Use:   "logs",
		Short: "Get logs across all instances of the capsule",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.logs),
	}
	capsuleLogs.Flags().BoolVarP(
		&follow, "follow", "f", false, "keep the connection open and read out logs as they are produced",
	)
	capsuleLogs.Flags().BoolVarP(
		&previousContainers, "previous-containers", "p", false,
		"Return logs from previous container terminations of the instance.",
	)
	capsuleLogs.Flags().StringVarP(&since, "since", "s", "1s", "do not show logs older than 'since'")
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

	var capsuleIDs []string

	if c.Scope.GetCurrentContext() == nil || c.Scope.GetCurrentContext().GetAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().List(ctx, &connect.Request[capsule_api.ListRequest]{
		Msg: &capsule_api.ListRequest{
			ProjectId: flags.GetProject(c.Scope),
		},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, c := range resp.Msg.GetCapsules() {
		if strings.HasPrefix(c.GetCapsuleId(), toComplete) {
			capsuleIDs = append(capsuleIDs, formatCapsule(c))
		}
	}

	if len(capsuleIDs) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return capsuleIDs, cobra.ShellCompDirectiveDefault
}

func formatCapsule(c *capsule_api.Capsule) string {
	age := time.Since(c.GetUpdatedAt().AsTime()).Truncate(time.Second).String()

	return fmt.Sprintf("%v\t (Updated At: %v)", c.GetCapsuleId(), age)
}

func (c *Cmd) persistentPreRunE(ctx context.Context, cmd *cobra.Command, _ []string) error {
	if _, ok := cmd.Annotations[auth.OmitCapsule]; ok {
		return nil
	}

	if capsule.CapsuleID != "" {
		return nil
	}

	name, err := capsule.SelectCapsule(ctx, c.Rig, c.Prompter, c.Scope)
	if err != nil {
		return err
	}

	capsule.CapsuleID = name
	return nil
}
