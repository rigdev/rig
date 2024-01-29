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
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/builddeploy"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/env"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/instance"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/jobs"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/mount"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/network"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/rollout"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/scale"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/errors"
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
	deleteCmd          bool
	previousContainers bool
)

var since string

type Cmd struct {
	fx.In

	Rig          rig.Client
	Cfg          *cmdconfig.Config
	DockerClient *client.Client
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Cfg = c.Cfg
	cmd.DockerClient = c.DockerClient
}

func Setup(parent *cobra.Command) {
	capsuleCmd := &cobra.Command{
		Use:   "capsule",
		Short: "Manage capsules",
		PersistentPreRunE: base.MakeInvokePreRunE(
			initCmd,
			func(ctx context.Context, cmd Cmd, c *cobra.Command, args []string) error {
				return cmd.persistentPreRunE(ctx, c, args)
			},
		),
	}
	capsuleCmd.PersistentFlags().StringVarP(&capsule.CapsuleID, "capsule-id", "c", "", "Id of the capsule")
	if err := capsuleCmd.RegisterFlagCompletionFunc(
		"capsule-id",
		base.CtxWrapCompletion(cmd.completions),
	); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	capsuleCreate := &cobra.Command{
		Use:   "create",
		Short: "Create a new capsule",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.create),
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
		RunE:  base.CtxWrap(cmd.abort),
	}
	capsuleCmd.AddCommand(capsuleAbort)

	capsuleDelete := &cobra.Command{
		Use:   "delete",
		Short: "Delete a capsule",
		Args:  cobra.NoArgs,
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
		},
		RunE: base.CtxWrap(cmd.delete),
	}
	capsuleCmd.AddCommand(capsuleDelete)

	capsuleGet := &cobra.Command{
		Use:               "get",
		Short:             "Get one or more capsules",
		PersistentPreRunE: base.PersistentPreRunE,
		Args:              cobra.NoArgs,
		Annotations: map[string]string{
			auth.OmitCapsule: "",
		},
		RunE: base.CtxWrap(cmd.get),
	}
	capsuleGet.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	capsuleGet.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	capsuleCmd.AddCommand(capsuleGet)

	capsuleCmdArgs := &cobra.Command{
		Use:   "cmd [some-command arg1 arg2]",
		Short: "Add command and arguments to the capsule",
		RunE:  base.CtxWrap(cmd.cmdArgs),
	}
	capsuleCmdArgs.Flags().BoolVarP(
		&forceDeploy,
		"force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes",
	)
	capsuleCmdArgs.Flags().BoolVarP(
		&deleteCmd,
		"delete", "d", false, "If set deletes the command and args from the capsule",
	)
	if err := capsuleCmdArgs.RegisterFlagCompletionFunc("force-deploy", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	capsuleCmd.AddCommand(capsuleCmdArgs)

	capsuleLogs := &cobra.Command{
		Use:   "logs",
		Short: "Get logs across all instances of the capsule",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.logs),
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

	scale.Setup(capsuleCmd)
	builddeploy.Setup(capsuleCmd)
	instance.Setup(capsuleCmd)
	network.Setup(capsuleCmd)
	rollout.Setup(capsuleCmd)
	env.Setup(capsuleCmd)
	mount.Setup(capsuleCmd)
	jobs.Setup(capsuleCmd)

	parent.AddCommand(capsuleCmd)
}

func (c *Cmd) completions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if err := base.Provide(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var capsuleIDs []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().List(ctx, &connect.Request[capsule_api.ListRequest]{
		Msg: &capsule_api.ListRequest{
			ProjectId: flags.GetProject(c.Cfg),
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
	var age string
	if c.GetCurrentRollout() == 0 {
		age = "-"
	} else {
		age = time.Since(c.GetUpdatedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (Rollout: %v, Updated At: %v)", c.GetCapsuleId(), c.GetCurrentRollout(), age)
}

func (c *Cmd) persistentPreRunE(ctx context.Context, cmd *cobra.Command, _ []string) error {
	if _, ok := cmd.Annotations[auth.OmitCapsule]; ok {
		return nil
	}

	if capsule.CapsuleID != "" {
		return nil
	}

	resp, err := c.Rig.Capsule().List(ctx, connect.NewRequest(&capsule_api.ListRequest{
		Pagination: &model.Pagination{},
		ProjectId:  flags.GetProject(c.Cfg),
	}))
	if err != nil {
		return err
	}

	var capsuleNames []string
	for _, c := range resp.Msg.GetCapsules() {
		capsuleNames = append(capsuleNames, c.GetCapsuleId())
	}

	if len(capsuleNames) == 0 {
		return errors.New("This project has no capsules. Create one, to get started")
	}

	_, name, err := common.PromptSelect("Capsule: ", capsuleNames, common.SelectFuzzyFilterOpt)
	if err != nil {
		return err
	}
	capsule.CapsuleID = name

	return nil
}
