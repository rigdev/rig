package build

import (
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
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
	image string
)

func Setup(parent *cobra.Command) *cobra.Command {
	build := &cobra.Command{
		Use:   "build",
		Short: "Manage builds of the capsule",
	}

	buildCreate := &cobra.Command{
		Use:               "create",
		Short:             "Create a new build with the given image",
		Args:              cobra.NoArgs,
		RunE:              base.Register(create),
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
		RunE:              base.Register(get),
		ValidArgsFunction: common.Complete(capsule.BuildCompletions, common.MaxArgsCompletionFilter(1)),
	}
	buildGet.Flags().IntVarP(&offset, "offset", "o", 0, "offset")
	buildGet.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	buildGet.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	buildGet.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	buildGet.RegisterFlagCompletionFunc("offset", common.NoCompletions)
	buildGet.RegisterFlagCompletionFunc("limit", common.NoCompletions)
	build.AddCommand(buildGet)

	parent.AddCommand(build)

	return build
}
