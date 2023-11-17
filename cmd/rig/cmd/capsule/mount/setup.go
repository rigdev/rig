package mount

import (
	"context"
	"fmt"
	"strings"
	"time"

	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	outputJSON  bool
	secret      bool
	forceDeploy bool
)

var (
	srcPath string
	dstPath string
)

type Cmd struct {
	fx.In

	Rig rig.Client
	Cfg *cmd_config.Config
}

func Setup(parent *cobra.Command) {
	mount := &cobra.Command{
		Use:   "mount",
		Short: "Manage config files mounts in the capsule",
	}

	mountGet := &cobra.Command{
		Use:   "get [mount-path]",
		Short: "Get one or multiple mounts",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(func(c Cmd) any { return c.get }),
		ValidArgsFunction: common.Complete(
			base.RegisterCompletion(func(c Cmd) any { return c.completions }),
			common.MaxArgsCompletionFilter(1),
		),
	}
	mountGet.Flags().StringVar(&dstPath, "download", "", "download the mount to specified path. If empty use current dir")
	mountGet.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	mountGet.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	mount.AddCommand(mountGet)

	mountSet := &cobra.Command{
		Use:               "set",
		Short:             "Mount a local configuration file in specified path the capsule",
		Args:              cobra.NoArgs,
		RunE:              base.Register(func(c Cmd) any { return c.set }),
		ValidArgsFunction: common.NoCompletions,
	}
	mountSet.Flags().StringVar(&srcPath, "src", "", "source path")
	mountSet.Flags().StringVar(&dstPath, "dst", "", "destination path")
	mountSet.Flags().BoolVarP(&secret, "secret", "s", false, "mount as secret")
	mountSet.Flags().BoolVarP(&forceDeploy, "force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes")
	mountSet.RegisterFlagCompletionFunc("dst", common.NoCompletions)
	mountSet.RegisterFlagCompletionFunc("src", common.NoCompletions)
	mount.AddCommand(mountSet)

	mountRemove := &cobra.Command{
		Use:   "remove [mount-path]",
		Short: "Remove a mount",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(func(c Cmd) any { return c.remove }),
		ValidArgsFunction: common.Complete(
			base.RegisterCompletion(func(c Cmd) any { return c.completions }),
			common.MaxArgsCompletionFilter(1),
		),
	}
	mountRemove.Flags().BoolVarP(&forceDeploy, "force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes")
	mountRemove.RegisterFlagCompletionFunc("force-deploy", common.BoolCompletions)
	mount.AddCommand(mountRemove)

	parent.AddCommand(mount)

}

func (c Cmd) completions(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if capsule.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	var paths []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	r, err := capsule.GetCurrentRollout(ctx, c.Rig)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, f := range r.GetConfig().GetConfigFiles() {
		if strings.HasPrefix(f.GetPath(), toComplete) {
			paths = append(paths, formatMount(f))
		}
	}

	if len(paths) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return paths, cobra.ShellCompDirectiveDefault
}

func formatMount(m *capsule_api.ConfigFile) string {
	var age string
	if m.GetUpdatedAt().AsTime().IsZero() {
		age = "-"
	} else {
		age = time.Since(m.GetUpdatedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (Age: %v)", m.GetPath(), age)
}
