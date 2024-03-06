package imagedeploy

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/docker/docker/client"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/image"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/cmd/rig/services/auth"
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
	setupImage(parent)
	setupDeploy(parent)
}

func setupImage(parent *cobra.Command) {
	image := &cobra.Command{
		Use:               "image",
		Short:             "Manage images of the capsule",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	imageCreate := &cobra.Command{
		Use:   "create",
		Short: "Create a new image with the given image",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.createImage),
	}
	imageCreate.Flags().StringVarP(&imageID, "image", "i", "", "image to use for the image")
	imageCreate.Flags().BoolVarP(&deploy, "deploy", "d", false, "deploy image after successful creation")
	imageCreate.Flags().BoolVarP(
		&forceDeploy, "force-deploy", "f", false, "force deploy. Aborting a deployment if one is in progress",
	)
	imageCreate.Flags().BoolVarP(
		&skipImageCheck, "skip-image-check", "s", false, "skip validating that the docker image exists",
	)
	imageCreate.Flags().BoolVarP(
		&remote, "remote", "r", false, "Rig will not look for the image locally but assumes it from a remote "+
			"registry. If not set, Rig will search locally and then remotely",
	)

	if err := imageCreate.RegisterFlagCompletionFunc("deploy", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := imageCreate.RegisterFlagCompletionFunc("force-deploy", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := imageCreate.RegisterFlagCompletionFunc("skip-image-check", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := imageCreate.RegisterFlagCompletionFunc("remote", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	image.AddCommand(imageCreate)

	imageGet := &cobra.Command{
		Use:   "get [image-id]",
		Short: "Get one or multiple images",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.CtxWrap(cmd.getBuild),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
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

func setupDeploy(parent *cobra.Command) {
	capsuleDeploy := &cobra.Command{
		Use:               "deploy",
		Short:             "Deploy the given image to a capsule",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
		Args:              cobra.NoArgs,
		RunE:              base.CtxWrap(cmd.deploy),
		Long: `Deploy either the given rig-image or docker image to a capsule.
If --image-id is given rig tries to find a matching existing rig-image to deploy.
If --image is given rig tries to create a new rig-image from the docker image (if it doesn't already exist)
Not both --image-id and --image can be given`,
	}
	capsuleDeploy.Flags().StringVarP(
		&imageID,
		"image", "i", "", "docker image to deploy. Will register the image in rig if it doesn't exist",
	)
	capsuleDeploy.Flags().BoolVarP(
		&remote, "remote", "r", false, "if --image is also given, Rig will assume the image is from a remote "+
			"registry. If not set, Rig will search locally and then remotely",
	)
	capsuleDeploy.Flags().BoolVarP(
		&forceDeploy, "force-deploy", "f", false, "force deploy. Aborting a rollout if one is in progress",
	)
	if err := capsuleDeploy.RegisterFlagCompletionFunc(
		"build-id",
		base.CtxWrapCompletion(cmd.completions),
	); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := capsuleDeploy.RegisterFlagCompletionFunc("force-deploy", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	parent.AddCommand(capsuleDeploy)
}

func (c *Cmd) completions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	if capsule.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	if err := base.Provide(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var buildIds []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Image().List(ctx, connect.NewRequest(
		&image.ListRequest{
			CapsuleId: capsule.CapsuleID,
			ProjectId: flags.GetProject(c.Cfg),
		}),
	)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, b := range resp.Msg.GetImages() {
		if strings.HasPrefix(b.GetImageId(), toComplete) {
			buildIds = append(buildIds, formatBuild(b))
		}
	}

	if len(buildIds) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return buildIds, cobra.ShellCompDirectiveDefault
}

func formatBuild(i *capsule_api.Image) string {
	var age string
	if i.GetCreatedAt().AsTime().IsZero() {
		age = "-"
	} else {
		age = time.Since(i.GetCreatedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (Age: %v)", i.GetImageId(), age)
}
