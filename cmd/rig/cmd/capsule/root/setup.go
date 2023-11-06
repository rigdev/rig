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
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/builddeploy"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/env"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/instance"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/mount"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/network"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/rollout"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/scale"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
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
)

var (
	command string
	args    []string
)

var omitCapsuleIDAnnotation = map[string]string{
	"OMIT_CAPSULE_ID": "true",
}

type Cmd struct {
	fx.In

	Ctx          context.Context
	Rig          rig.Client
	Cfg          *cmd_config.Config
	DockerClient *client.Client

	Scale       scale.Cmd
	BuildDeploy builddeploy.Cmd
	Instance    instance.Cmd
	Network     network.Cmd
	Rollout     rollout.Cmd
	Env         env.Cmd
	Mount       mount.Cmd
}

func (c Cmd) Setup(parent *cobra.Command) {
	capsuleCmd := &cobra.Command{
		Use:   "capsule",
		Short: "Manage capsules",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if parent.PersistentPreRunE != nil {
				if err := parent.PersistentPreRunE(cmd, args); err != nil {
					return err
				}
			}
			if cmd.Annotations["OMIT_CAPSULE_ID"] != "" {
				return nil
			}

			if capsule.CapsuleID != "" {
				return nil
			}

			resp, err := c.Rig.Capsule().List(c.Ctx, connect.NewRequest(&capsule_api.ListRequest{
				Pagination: &model.Pagination{},
			}))
			if err != nil {
				return err
			}

			var capsuleNames []string
			for _, c := range resp.Msg.GetCapsules() {
				capsuleNames = append(capsuleNames, c.GetCapsuleId())
			}

			_, name, err := common.PromptSelect("Capsule: ", capsuleNames, common.SelectFuzzyFilterOpt)
			if err != nil {
				return err
			}
			capsule.CapsuleID = name

			return nil
		},
	}
	capsuleCmd.PersistentFlags().StringVarP(&capsule.CapsuleID, "capsule-id", "c", "", "Id of the capsule")
	capsuleCmd.RegisterFlagCompletionFunc("capsule-id", c.completions)

	capsuleCreate := &cobra.Command{
		Use:               "create",
		Short:             "Create a new capsule",
		Args:              cobra.NoArgs,
		RunE:              c.create,
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
		RunE:              c.abort,
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleCmd.AddCommand(capsuleAbort)

	capsuleDelete := &cobra.Command{
		Use:               "delete",
		Short:             "Delete a capsule",
		Args:              cobra.NoArgs,
		RunE:              c.delete,
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleCmd.AddCommand(capsuleDelete)

	capsuleGet := &cobra.Command{
		Use:               "get",
		Short:             "Get one or more capsules",
		Args:              cobra.NoArgs,
		Annotations:       omitCapsuleIDAnnotation,
		RunE:              c.get,
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
		RunE:              c.config,
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

	c.Scale.Setup(capsuleCmd)
	c.BuildDeploy.Setup(capsuleCmd)
	c.Instance.Setup(capsuleCmd)
	c.Network.Setup(capsuleCmd)
	c.Rollout.Setup(capsuleCmd)
	c.Env.Setup(capsuleCmd)
	c.Mount.Setup(capsuleCmd)

	parent.AddCommand(capsuleCmd)
}

func (c Cmd) completions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var capsuleIDs []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().List(c.Ctx, &connect.Request[capsule_api.ListRequest]{
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
