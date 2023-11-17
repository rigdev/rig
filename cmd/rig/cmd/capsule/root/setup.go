package root

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
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
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/mount"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/network"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/rollout"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/scale"
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
	interactive bool
	outputJSON  bool
	forceDeploy bool
	follow      bool
)

var (
	command string
	args    []string
	since   string
)

var omitCapsuleIDAnnotation = map[string]string{
	"OMIT_CAPSULE_ID": "true",
}

type Cmd struct {
	fx.In

	Rig          rig.Client
	Cfg          *cmd_config.Config
	DockerClient *client.Client
}

func Setup(parent *cobra.Command) {
	capsuleCmd := &cobra.Command{
		Use:   "capsule",
		Short: "Manage capsules",
		PersistentPreRunE: base.Register(func(c Cmd) any {
			return c.persistentPreRunE
		}),
	}
	capsuleCmd.PersistentFlags().StringVarP(&capsule.CapsuleID, "capsule-id", "c", "", "Id of the capsule")
	capsuleCmd.RegisterFlagCompletionFunc(
		"capsule-id",
		base.RegisterCompletion(func(c Cmd) any { return c.completions }),
	)

	capsuleCreate := &cobra.Command{
		Use:               "create",
		Short:             "Create a new capsule",
		Args:              cobra.NoArgs,
		RunE:              base.Register(func(c Cmd) any { return c.create }),
		Annotations:       omitCapsuleIDAnnotation,
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleCreate.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactive mode")
	capsuleCreate.Flags().BoolVarP(&forceDeploy, "force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes")
	capsuleCreate.RegisterFlagCompletionFunc("interactive", common.BoolCompletions)
	capsuleCreate.RegisterFlagCompletionFunc("force-deploy", common.BoolCompletions)
	capsuleCmd.AddCommand(capsuleCreate)

	capsuleAbort := &cobra.Command{
		Use:               "abort",
		Short:             "Abort the current rollout. This will leave the capsule in a undefined state",
		Args:              cobra.NoArgs,
		RunE:              base.Register(func(c Cmd) any { return c.abort }),
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleCmd.AddCommand(capsuleAbort)

	capsuleDelete := &cobra.Command{
		Use:               "delete",
		Short:             "Delete a capsule",
		Args:              cobra.NoArgs,
		RunE:              base.Register(func(c Cmd) any { return c.delete }),
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleCmd.AddCommand(capsuleDelete)

	capsuleGet := &cobra.Command{
		Use:               "get",
		Short:             "Get one or more capsules",
		Args:              cobra.NoArgs,
		Annotations:       omitCapsuleIDAnnotation,
		RunE:              base.Register(func(c Cmd) any { return c.get }),
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleGet.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	capsuleGet.Flags().IntVarP(&offset, "offset", "o", 0, "offset for pagination")
	capsuleGet.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	capsuleGet.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	capsuleGet.RegisterFlagCompletionFunc("offset", common.NoCompletions)
	capsuleGet.RegisterFlagCompletionFunc("limit", common.NoCompletions)
	capsuleCmd.AddCommand(capsuleGet)

	capsuleConfig := &cobra.Command{
		Use:               "config",
		Short:             "Configure the capsule",
		Args:              cobra.NoArgs,
		RunE:              base.Register(func(c Cmd) any { return c.config }),
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleConfig.Flags().Bool("auto-add-service-account", false, "automatically add the rig service account to the capsule")
	capsuleConfig.Flags().StringVar(&command, "cmd", "", "Container CMD to run")
	capsuleConfig.Flags().StringSliceVar(&args, "args", []string{}, "Container CMD args")
	capsuleConfig.Flags().BoolVarP(&forceDeploy, "force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes")
	capsuleConfig.RegisterFlagCompletionFunc("force-deploy", common.BoolCompletions)
	capsuleConfig.RegisterFlagCompletionFunc("auto-add-service-account", common.BoolCompletions)
	capsuleConfig.RegisterFlagCompletionFunc("cmd", common.NoCompletions)
	capsuleConfig.RegisterFlagCompletionFunc("args", common.NoCompletions)
	capsuleCmd.AddCommand(capsuleConfig)

	capsuleLogs := &cobra.Command{
		Use:               "logs",
		Short:             "Get logs across all instances of the capsule",
		Args:              cobra.NoArgs,
		RunE:              base.Register(func(c Cmd) any { return c.logs }),
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleLogs.Flags().BoolVarP(&follow, "follow", "f", false, "keep the connection open and read out logs as they are produced")
	capsuleLogs.Flags().StringVarP(&since, "since", "s", "1s", "do not show logs older than 'since'")
	capsuleCmd.AddCommand(capsuleLogs)

	scale.Setup(capsuleCmd)
	builddeploy.Setup(capsuleCmd)
	instance.Setup(capsuleCmd)
	network.Setup(capsuleCmd)
	rollout.Setup(capsuleCmd)
	env.Setup(capsuleCmd)
	mount.Setup(capsuleCmd)

	parent.AddCommand(capsuleCmd)
}

func (c Cmd) completions(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var capsuleIDs []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().List(ctx, &connect.Request[capsule_api.ListRequest]{
		Msg: &capsule_api.ListRequest{},
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

func (c Cmd) persistentPreRunE(ctx context.Context, cmd *cobra.Command, args []string) error {
	base.ExecutePersistentPreRunERecursively(cmd, args)
	if cmd.Annotations["OMIT_CAPSULE_ID"] != "" {
		return nil
	}

	if capsule.CapsuleID != "" {
		return nil
	}

	resp, err := c.Rig.Capsule().List(ctx, connect.NewRequest(&capsule_api.ListRequest{
		Pagination: &model.Pagination{},
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
