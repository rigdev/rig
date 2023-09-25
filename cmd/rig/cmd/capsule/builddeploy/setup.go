package builddeploy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/docker/docker/client"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	outputJSON     bool
	deploy         bool
	skipImageCheck bool
	remote         bool
)

var (
	image   string
	buildID string
)

type Cmd struct {
	fx.In

	Ctx          context.Context
	Rig          rig.Client
	Cfg          *cmd_config.Config
	DockerClient *client.Client
}

func (c Cmd) Setup(parent *cobra.Command) {
	c.setupBuild(parent)
	c.setupDeploy(parent)
}

func (c Cmd) setupBuild(parent *cobra.Command) {
	build := &cobra.Command{
		Use:   "build",
		Short: "Manage builds of the capsule",
	}

	buildCreate := &cobra.Command{
		Use:               "create",
		Short:             "Create a new build with the given image",
		Args:              cobra.NoArgs,
		RunE:              c.createBuild,
		ValidArgsFunction: common.NoCompletions,
	}
	buildCreate.Flags().StringVarP(&image, "image", "i", "", "image to use for the build")
	buildCreate.Flags().BoolVarP(&deploy, "deploy", "d", false, "deploy build after successful creation")
	buildCreate.Flags().BoolVarP(&skipImageCheck, "skip-image-check", "s", false, "skip validating that the docker image exists")
	buildCreate.Flags().BoolVarP(&remote, "remote", "r", false, "Rig will not look for the image locally but assumes it from a remote registry. If not set, Rig will search locally and then remotely")

	buildCreate.RegisterFlagCompletionFunc("image", common.NoCompletions)
	buildCreate.RegisterFlagCompletionFunc("deploy", common.BoolCompletions)
	buildCreate.RegisterFlagCompletionFunc("skip-image-check", common.BoolCompletions)
	buildCreate.RegisterFlagCompletionFunc("remote", common.BoolCompletions)
	build.AddCommand(buildCreate)

	buildGet := &cobra.Command{
		Use:               "get [build-id]",
		Short:             "Get one or multiple builds",
		Args:              cobra.MaximumNArgs(1),
		RunE:              c.getBuild,
		ValidArgsFunction: common.Complete(c.completions, common.MaxArgsCompletionFilter(1)),
	}
	buildGet.Flags().IntVarP(&offset, "offset", "o", 0, "offset")
	buildGet.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	buildGet.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	buildGet.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	buildGet.RegisterFlagCompletionFunc("offset", common.NoCompletions)
	buildGet.RegisterFlagCompletionFunc("limit", common.NoCompletions)
	build.AddCommand(buildGet)

	parent.AddCommand(build)
}

func (c Cmd) setupDeploy(parent *cobra.Command) {
	capsuleDeploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the given build to a capsule",
		Args:  cobra.NoArgs,
		RunE:  c.deploy,
		Long: `Deploy either the given rig-build or docker image to a capsule.
If --build-id is given rig tries to find a matching existing rig-build to deploy.
If --image is given rig tries to create a new rig-build from the docker image (if it doesn't already exist)
Not both --build-id and --image can be given`,
	}
	capsuleDeploy.Flags().StringVarP(&buildID, "build-id", "b", "", "rig build id to deploy")
	capsuleDeploy.Flags().StringVarP(&image, "image", "i", "", "docker image to deploy. Will create a new rig-build from the image if it doesn't exist")
	capsuleDeploy.Flags().BoolVarP(&remote, "remote", "r", false, "if --image is also given, Rig will assume the image is from a remote registry. If not set, Rig will search locally and then remotely")
	capsuleDeploy.RegisterFlagCompletionFunc("build-id", c.completions)
	capsuleDeploy.RegisterFlagCompletionFunc("image", common.NoCompletions)
	capsuleDeploy.RegisterFlagCompletionFunc("remote", common.BoolCompletions)

	parent.AddCommand(capsuleDeploy)
}

func (c Cmd) completions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	if capsule.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	var buildIds []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().ListBuilds(c.Ctx, &connect.Request[capsule_api.ListBuildsRequest]{
		Msg: &capsule_api.ListBuildsRequest{
			CapsuleId: capsule.CapsuleID,
		},
	})
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
