package capsule

import (
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
)

var (
	offset int
	limit  int
)

var (
	interactive bool
	outputJSON  bool
	remote      bool
)

var (
	CapsuleID string
	image     string
	buildID   string
	command   string
)

var (
	args []string
)

var omitCapsuleIDAnnotation = map[string]string{
	"OMIT_CAPSULE_ID": "true",
}

func Setup(parent *cobra.Command) *cobra.Command {
	capsuleCmd := &cobra.Command{
		Use:   "capsule",
		Short: "Manage capsules",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Annotations["OMIT_CAPSULE_ID"] != "" {
				return nil
			}

			if CapsuleID == "" {
				var err error
				CapsuleID, err = common.PromptInput("Capsule id:", common.ValidateNonEmptyOpt)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	capsuleCmd.PersistentFlags().StringVarP(&CapsuleID, "capsule-id", "c", "", "Id of the capsule")
	capsuleCmd.RegisterFlagCompletionFunc("capsule-id", capsuleCompletions)

	capsuleCreate := &cobra.Command{
		Use:               "create",
		Short:             "Create a new capsule",
		Args:              cobra.NoArgs,
		RunE:              base.Register(create),
		Annotations:       omitCapsuleIDAnnotation,
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleCreate.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactive mode")
	capsuleCreate.RegisterFlagCompletionFunc("interactive", common.BoolCompletions)

	capsuleCmd.AddCommand(capsuleCreate)

	capsuleDeploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the given build to a capsule",
		Args:  cobra.NoArgs,
		RunE:  base.Register(deploy),
		Long: `Deploy either the given rig-build or docker image to a capsule.
If --build-id is given rig tries to find a matching existing rig-build to deploy.
If --image is given rig tries to create a new rig-build from the docker image (if it doesn't already exist)
Not both --build-id and --image can be given`,
	}
	capsuleDeploy.Flags().StringVarP(&buildID, "build-id", "b", "", "rig build id to deploy")
	capsuleDeploy.Flags().StringVarP(&image, "image", "i", "", "docker image to deploy. Will create a new rig-build from the image if it doesn't exist")
	capsuleDeploy.Flags().BoolVarP(&remote, "remote", "r", false, "if --image is also given, Rig will assume the image is from a remote registry. If not set, Rig will search locally and then remotely")
	capsuleDeploy.RegisterFlagCompletionFunc("build-id", BuildCompletions)
	capsuleDeploy.RegisterFlagCompletionFunc("image", common.NoCompletions)
	capsuleDeploy.RegisterFlagCompletionFunc("remote", common.BoolCompletions)
	capsuleCmd.AddCommand(capsuleDeploy)

	capsuleAbort := &cobra.Command{
		Use:               "abort",
		Short:             "Abort the current rollout. This will leave the capsule in a undefined state",
		Args:              cobra.NoArgs,
		RunE:              base.Register(abort),
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleCmd.AddCommand(capsuleAbort)

	capsuleDelete := &cobra.Command{
		Use:               "delete",
		Short:             "Delete a capsule",
		Args:              cobra.NoArgs,
		RunE:              base.Register(delete),
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleCmd.AddCommand(capsuleDelete)

	capsuleGet := &cobra.Command{
		Use:               "get",
		Short:             "Get one or more capsules",
		Args:              cobra.NoArgs,
		Annotations:       omitCapsuleIDAnnotation,
		RunE:              base.Register(get),
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
		RunE:              base.Register(config),
		ValidArgsFunction: common.NoCompletions,
	}
	capsuleConfig.Flags().Bool("auto-add-service-account", false, "automatically add the rig service account to the capsule")
	capsuleConfig.Flags().StringVar(&command, "cmd", "", "Container CMD to run")
	capsuleConfig.Flags().StringSliceVar(&args, "args", []string{}, "Container CMD args")
	capsuleConfig.RegisterFlagCompletionFunc("auto-add-service-account", common.BoolCompletions)
	capsuleConfig.RegisterFlagCompletionFunc("cmd", common.NoCompletions)
	capsuleConfig.RegisterFlagCompletionFunc("args", common.NoCompletions)
	capsuleCmd.AddCommand(capsuleConfig)

	parent.AddCommand(capsuleCmd)

	return capsuleCmd
}
