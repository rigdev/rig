package deploy

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
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	skipImageCheck             bool
	remote                     bool
	noWait                     bool
	environmentVariables       map[string]string
	removeEnvironmentVariables []string
	environmentSources         []string
	removeEnvironmentSources   []string
	annotations                map[string]string
	removeAnnotations          []string
	replicas                   int
	configFiles                []string
	removeConfigFiles          []string
)

var imageID string

type Cmd struct {
	fx.In

	Rig          rig.Client
	Cfg          *cmdconfig.Config
	DockerClient *client.Client
	Interactive  cli.Interactive
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd = c
}

func Setup(parent *cobra.Command) {
	capsuleDeploy := &cobra.Command{
		Use:               "deploy [capsule] [flags] [-- command]",
		Short:             "Deploy changes to a capsule",
		PersistentPreRunE: cli.MakeInvokePreRunE(initCmd),
		RunE:              cli.CtxWrap(cmd.deploy),
		Long: `Deploy a number of changes to a Capsule.

All the changes given will be deployed as one rollout, then waiting for the rollout to complete.
Use '--no-wait' to skip this.

If --image is given, rig creates a new reference to the docker image if it doesn't already exist`,
	}
	capsuleDeploy.Flags().BoolVar(&noWait, "no-wait", false, "skip waiting for the changes to be applied.")

	capsuleDeploy.Flags().StringToStringVarP(
		&environmentVariables,
		"set-env-var", "e", nil,
		"environment variables to add to the Capsule of the format `key=value`",
	)
	capsuleDeploy.Flags().StringSliceVar(
		&removeEnvironmentVariables,
		"rm-env-var", nil,
		"environment variables to remove from the Capsule",
	)

	capsuleDeploy.Flags().IntVarP(
		&replicas,
		"replicas", "r", 0,
		"replicas of the Capsule to run. If Autoscaler is enabled, this will change the minimum number of replicas "+
			"for the Capsule",
	)
	capsuleDeploy.Flags().StringToStringVarP(
		&annotations,
		"set-annotation", "A", nil,
		"annotations to add to the Capsule of the format `key=value`",
	)
	capsuleDeploy.Flags().StringSliceVar(
		&removeAnnotations,
		"rm-annotation", nil,
		"annotation to remove from the Capsule",
	)

	capsuleDeploy.Flags().StringSliceVar(
		&environmentSources,
		"set-env-source", nil,
		"environment source references to set on the Capsule. Must be of the format `[ConfigMap|Secret]/name`, "+
			"e.g. `Secret/my-secret`",
	)
	capsuleDeploy.Flags().StringSliceVar(
		&removeEnvironmentSources,
		"rm-env-source", nil,
		"environment source references to remove from the Capsule. Must be of the format `[ConfigMap|Secret]/name`, "+
			"e.g. `Secret/my-secret`",
	)
	capsuleDeploy.Flags().StringVarP(
		&imageID,
		"image", "i", "", "container image to deploy. Will register the image in rig if it doesn't exist",
	)
	capsuleDeploy.Flags().BoolVar(
		&remote, "remote", false, "if --image is also given, Rig will assume the image is from a remote "+
			"registry. If not set, Rig will search locally and then remotely",
	)
	capsuleDeploy.Flags().StringArrayVar(
		&configFiles, "set-config-file", nil,
		"config files to set in the capsule, adding if not already exists. Must be a mapping from "+
			"`path=<container-path>,src=<file-path>,[options]`, where `file-path` must be a local file and `container-path` "+
			"is an absolute path within the container. Options can be `secret`, which "+
			"would create the resource as a Kubernetes Secret.",
	)
	capsuleDeploy.Flags().StringSliceVar(
		&removeConfigFiles, "rm-config-file", nil, "config files to remove from the capsule. Must be an absolute path "+
			"of the config-file within the container",
	)

	if err := capsuleDeploy.RegisterFlagCompletionFunc(
		"image",
		cli.CtxWrapCompletion(cmd.completions),
	); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	parent.AddCommand(capsuleDeploy)

	// Add as top-level command as well.
	parent.Parent().AddCommand(capsuleDeploy)
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

func formatImage(i *capsule_api.Image) string {
	var age string
	if i.GetCreatedAt().AsTime().IsZero() {
		age = "-"
	} else {
		age = time.Since(i.GetCreatedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (Age: %v)", i.GetImageId(), age)
}
