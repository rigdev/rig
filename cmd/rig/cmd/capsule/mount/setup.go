package mount

import (
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

var (
	outputJSON bool
)

var (
	srcPath string
	dstPath string
	// path string
)

func Setup(parent *cobra.Command) *cobra.Command {
	mount := &cobra.Command{
		Use:   "mount",
		Short: "Manage config files mounts in the capsule",
	}

	mountGet := &cobra.Command{
		Use:               "get [mount-path]",
		Short:             "Get one or multiple mounts",
		Args:              cobra.MaximumNArgs(1),
		RunE:              base.Register(get),
		ValidArgsFunction: common.Complete(capsule.MountCompletions, common.MaxArgsCompletionFilter(1)),
	}
	mountGet.Flags().StringVar(&dstPath, "download", "", "download the mount to specified path. If empty use current dir")
	mountGet.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	mountGet.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	mount.AddCommand(mountGet)

	mountSet := &cobra.Command{
		Use:               "set",
		Short:             "Mount a local configuration file in specified path the capsule",
		Args:              cobra.NoArgs,
		RunE:              base.Register(set),
		ValidArgsFunction: common.NoCompletions,
	}
	mountSet.Flags().StringVar(&srcPath, "src", "", "source path")
	mountSet.Flags().StringVar(&dstPath, "dst", "", "destination path")
	mountSet.RegisterFlagCompletionFunc("dst", common.NoCompletions)
	mount.AddCommand(mountSet)

	mountRemove := &cobra.Command{
		Use:               "remove [mount-path]",
		Short:             "Remove a mount",
		Args:              cobra.MaximumNArgs(1),
		RunE:              base.Register(remove),
		ValidArgsFunction: common.Complete(capsule.MountCompletions, common.MaxArgsCompletionFilter(1)),
	}
	mount.AddCommand(mountRemove)

	parent.AddCommand(mount)
	return mount

}
