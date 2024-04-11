package image

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
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
	deploy         bool
	skipImageCheck bool
	remote         bool
	forceDeploy    bool
)

var imageID string

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
	image := &cobra.Command{
		Use:               "image",
		Short:             "Manage images of the capsule",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		GroupID:           capsule.DeploymentGroupID,
	}

	imageAdd := &cobra.Command{
		Use:   "add [capsule]",
		Short: "Add a new container image",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.addImage),
	}
	imageAdd.Flags().StringVarP(&imageID, "image", "i", "", "image to use for the image")
	imageAdd.Flags().BoolVarP(&deploy, "deploy", "d", false, "deploy image after successful creation")
	imageAdd.Flags().BoolVarP(
		&forceDeploy, "force-deploy", "f", false, "force deploy. Aborting a deployment if one is in progress",
	)
	imageAdd.Flags().BoolVarP(
		&skipImageCheck, "skip-image-check", "s", false, "skip validating that the docker image exists",
	)
	imageAdd.Flags().BoolVarP(
		&remote, "remote", "r", false, "Rig will not look for the image locally but assumes it from a remote "+
			"registry. If not set, Rig will search locally and then remotely",
	)

	if err := imageAdd.RegisterFlagCompletionFunc("deploy", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := imageAdd.RegisterFlagCompletionFunc("force-deploy", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := imageAdd.RegisterFlagCompletionFunc("skip-image-check", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := imageAdd.RegisterFlagCompletionFunc("remote", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	image.AddCommand(imageAdd)

	imageGet := &cobra.Command{
		Use:   "list [capsule]",
		Short: "Get one or multiple images",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.list),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
		},
	}
	imageGet.Flags().IntVar(&offset, "offset", 0, "offset")
	imageGet.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	image.AddCommand(imageGet)

	parent.AddCommand(image)
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
