package image

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/docker/docker/client"
	"github.com/rigdev/rig-go-api/api/v1/image"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/cli"
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
	image := &cobra.Command{
		Use:               "image",
		Short:             "Manage images of the capsule",
		PersistentPreRunE: cli.MakeInvokePreRunE(initCmd),
	}

	imageCreate := &cobra.Command{
		Use:   "create",
		Short: "Create a new image with the given image",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.createImage),
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
		RunE:  cli.CtxWrap(cmd.getBuild),
		ValidArgsFunction: common.Complete(
			cli.CtxWrapCompletion(cmd.completions),
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

	if err := cli.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var imageIDs []string

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
			imageIDs = append(imageIDs, formatImage(b))
		}
	}

	if len(imageIDs) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return imageIDs, cobra.ShellCompDirectiveDefault
}
