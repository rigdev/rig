package builddeploy

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/docker/docker/client"
	"github.com/rigdev/rig-go-api/api/v1/build"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
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

var (
	image   string
	buildID string
)

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
	setupBuild(parent)
	setupDeploy(parent)
}

func setupBuild(parent *cobra.Command) {
	build := &cobra.Command{
		Use:               "build",
		Short:             "Manage builds of the capsule",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	buildCreate := &cobra.Command{
		Use:   "create",
		Short: "Create a new build with the given image",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.createBuild),
	}
	buildCreate.Flags().StringVarP(&image, "image", "i", "", "image to use for the build")
	buildCreate.Flags().BoolVarP(&deploy, "deploy", "d", false, "deploy build after successful creation")
	buildCreate.Flags().BoolVarP(
		&forceDeploy, "force-deploy", "f", false, "force deploy. Aborting a deployment if one is in progress",
	)
	buildCreate.Flags().BoolVarP(
		&skipImageCheck, "skip-image-check", "s", false, "skip validating that the docker image exists",
	)
	buildCreate.Flags().BoolVarP(
		&remote, "remote", "r", false, "Rig will not look for the image locally but assumes it from a remote "+
			"registry. If not set, Rig will search locally and then remotely",
	)

	if err := buildCreate.RegisterFlagCompletionFunc("deploy", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := buildCreate.RegisterFlagCompletionFunc("force-deploy", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := buildCreate.RegisterFlagCompletionFunc("skip-image-check", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := buildCreate.RegisterFlagCompletionFunc("remote", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	build.AddCommand(buildCreate)

	buildGet := &cobra.Command{
		Use:   "get [build-id]",
		Short: "Get one or multiple builds",
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
	buildGet.Flags().IntVar(&offset, "offset", 0, "offset")
	buildGet.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	build.AddCommand(buildGet)

	parent.AddCommand(build)
}

func setupDeploy(parent *cobra.Command) {
	capsuleDeploy := &cobra.Command{
		Use:               "deploy",
		Short:             "Deploy the given build to a capsule",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
		Args:              cobra.NoArgs,
		RunE:              base.CtxWrap(cmd.deploy),
		Long: `Deploy either the given rig-build or docker image to a capsule.
If --build-id is given rig tries to find a matching existing rig-build to deploy.
If --image is given rig tries to create a new rig-build from the docker image (if it doesn't already exist)
Not both --build-id and --image can be given`,
	}
	capsuleDeploy.Flags().StringVarP(&buildID, "build-id", "b", "", "rig build id to deploy")
	capsuleDeploy.Flags().StringVarP(
		&image,
		"image", "i", "", "docker image to deploy. Will create a new rig-build from the image if it doesn't exist",
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

	resp, err := c.Rig.Build().List(ctx, connect.NewRequest(
		&build.ListRequest{
			CapsuleId: capsule.CapsuleID,
			ProjectId: c.Cfg.GetProject(),
		}),
	)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, b := range resp.Msg.GetBuilds() {
		if strings.HasPrefix(b.GetBuildId(), toComplete) {
			buildIds = append(buildIds, formatBuild(b))
		}
	}

	if len(buildIds) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return buildIds, cobra.ShellCompDirectiveDefault
}

func formatBuild(b *capsule_api.Build) string {
	var age string
	if b.GetCreatedAt().AsTime().IsZero() {
		age = "-"
	} else {
		age = time.Since(b.GetCreatedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (Age: %v)", b.GetBuildId(), age)
}
