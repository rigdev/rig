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
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	// Configuration
	environmentVariables       map[string]string
	removeEnvironmentVariables []string
	environmentSources         []string
	removeEnvironmentSources   []string
	annotations                map[string]string
	removeAnnotations          []string
	replicas                   int
	configFiles                []string
	removeConfigFiles          []string
	networkInterfaces          []string
	removeNetworkInterfaces    []string
	file                       string

	// Other
	skipImageCheck     bool
	remote             bool
	noWait             bool
	forceOverride      bool
	currentRolloutID   uint64
	currentFingerprint string
	timeout            time.Duration
	noRollback         bool
	prBranchName       string
	dry                bool
	currentEnvRollouts string
)

var imageID string

type Cmd struct {
	fx.In

	Rig          rig.Client
	DockerClient *client.Client
	Scope        scope.Scope
	Prompter     common.Prompter
	Scheme       *runtime.Scheme
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	baseDeploy := cobra.Command{
		Use:   "deploy [capsule] [flags] [-- command]",
		Short: "Deploy changes to a capsule",
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			common.MaxArgsCompletionFilter(1)),
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		RunE:              cli.CtxWrap(cmd.deploy),
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
			auth.OmitProject:     "",
		},
		Long: `Deploy a number of changes to a Capsule.

All the changes given will be deployed as one rollout, then waiting for the rollout to complete.
Use '--no-wait' to skip this.

If --image is given, rig creates a new reference to the docker image if it doesn't already exist

If the capsule is configured to require a pull request, the deploy will be executed as a proposal,not a direct rollout.
In that case you must supply branch name in --pr-branch.`,
	}
	baseDeploy.Flags().BoolVar(&noWait, "no-wait", false, "skip waiting for the changes to be applied.")
	setupConfigurationFlags(&baseDeploy)

	baseDeploy.Flags().BoolVar(&forceOverride, "force-override", false,
		"by default, existing objects will be kept in favor of overriding them."+
			"To force the override of resources, set this flag to true."+
			"An example of this use-case is a migration step, where resource created by a previous toolchain e.g."+
			"based on Helm charts, are to be replaced and instead be created by the Rig operator."+
			"While the override is irreversible, this flag is not \"sticky\" and must be set by each"+
			"deploy that should use this behavior.",
	)
	baseDeploy.Flags().Uint64Var(
		&currentRolloutID, "current-rollout", 0,
		"if set, will verify that the current rollout is the one given.",
	)
	baseDeploy.Flags().DurationVarP(
		&timeout, "timeout", "t", 0,
		"timeout for when the deploy command should terminate."+
			" Unless --no-rollback is configured, this will result in a rollback.",
	)
	baseDeploy.Flags().BoolVar(
		&noRollback, "no-rollback", false,
		"disable automatic rollback if the change was unsuccessful.",
	)
	baseDeploy.Flags().StringVar(
		&prBranchName, "pr-branch", "",
		"if set will create a proposal pull request with the given branch name instead of a direct rollout.",
	)
	baseDeploy.Flags().BoolVarP(
		&dry, "dry", "d", false,
		"if set, will not apply the change but display the diff with the current capsule spec.",
	)

	if err := baseDeploy.RegisterFlagCompletionFunc(
		"image",
		cli.HackCtxWrapCompletion(cmd.imageCompletions, s),
	); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	capsuleDeploy := baseDeploy
	capsuleDeploy.GroupID = capsule.DeploymentGroupID
	parent.AddCommand(&capsuleDeploy)

	capsuleSetDeploy := &cobra.Command{
		Use:   "deploy-set [capsule] [flags] [-- command]",
		Short: "Deploy changes to a capsule set",
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			common.MaxArgsCompletionFilter(1)),
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		RunE:              cli.CtxWrap(cmd.deploySet),
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
		},
	}
	setupConfigurationFlags(capsuleSetDeploy)
	capsuleSetDeploy.Flags().StringVar(
		&currentEnvRollouts, "current-env-rollouts", "",
		"A comma seperated list of environtment, rolloutID mappings of the form 'env1:id1,env2:id2'."+
			" If set, will verify that the given environments current rollouts are as given.",
	)
	capsuleSetDeploy.Flags().BoolVar(&noWait, "no-wait", false, "skip waiting for the changes to be applied.")
	capsuleSetDeploy.Flags().DurationVarP(
		&timeout, "timeout", "t", 0,
		"timeout for when the deploy command should terminate."+
			" Unless --no-rollback is configured, this will result in a rollback.",
	)
	capsuleSetDeploy.Flags().StringVar(
		&prBranchName, "pr-branch", "",
		"if set will create a proposal pull request with the given branch name instead of a direct rollout.",
	)
	// capsuleSetDeploy.Flags().BoolVarP(
	// 	&dry, "dry", "d", false,
	// 	"if set, will not apply the change but display the diff with the current capsule spec.",
	// )

	// parent.AddCommand(capsuleSetDeploy) TODO Enable when ready!

	// Add as top-level command as well.
	rootDeploy := baseDeploy
	rootDeploy.GroupID = common.CapsuleGroupID
	parent.Parent().AddCommand(&rootDeploy)
}

func setupConfigurationFlags(c *cobra.Command) {
	c.Flags().StringToStringVarP(
		&environmentVariables,
		"set-env-var", "e", nil,
		"environment variables to add to the Capsule of the format `key=value`",
	)
	c.Flags().StringSliceVar(
		&removeEnvironmentVariables,
		"rm-env-var", nil,
		"environment variables to remove from the Capsule",
	)

	c.Flags().IntVarP(
		&replicas,
		"replicas", "r", 0,
		"replicas of the Capsule to run. If Autoscaler is enabled, this will change the minimum number of replicas "+
			"for the Capsule",
	)
	c.Flags().StringToStringVarP(
		&annotations,
		"set-annotation", "A", nil,
		"annotations to add to the Capsule of the format `key=value`",
	)
	c.Flags().StringSliceVar(
		&removeAnnotations,
		"rm-annotation", nil,
		"annotation to remove from the Capsule",
	)

	c.Flags().StringSliceVar(
		&environmentSources,
		"set-env-source", nil,
		"environment source references to set on the Capsule. Must be of the format `[ConfigMap|Secret]/name`, "+
			"e.g. `Secret/my-secret`",
	)
	c.Flags().StringSliceVar(
		&removeEnvironmentSources,
		"rm-env-source", nil,
		"environment source references to remove from the Capsule. Must be of the format `[ConfigMap|Secret]/name`, "+
			"e.g. `Secret/my-secret`",
	)
	c.Flags().StringVarP(
		&imageID,
		"image", "i", "", "container image to deploy. Will register the image in rig if it doesn't exist",
	)
	c.Flags().BoolVar(
		&remote, "remote", false, "if --image is also given, Rig will assume the image is from a remote "+
			"registry. If not set, Rig will search locally and then remotely",
	)
	c.Flags().StringArrayVar(
		&configFiles, "set-config-file", nil,
		"config files to set in the capsule, adding if not already exists. Must be a mapping from "+
			"`path=<container-path>,src=<file-path>,[options]`, where `file-path` must be a local file and `container-path` "+
			"is an absolute path within the container. Options can be `secret`, which "+
			"would create the resource as a Kubernetes Secret.",
	)
	c.Flags().StringSliceVar(
		&removeConfigFiles, "rm-config-file", nil, "config files to remove from the capsule. Must be an absolute path "+
			"of the config-file within the container",
	)
	c.Flags().StringSliceVar(&networkInterfaces, "set-network-interface", nil,
		"create or update the network interface. The argument is a file from where the network interface "+
			"can be read. The Network Interface must have both a name and a port.")
	c.Flags().StringSliceVar(&removeNetworkInterfaces, "rm-network-interface", nil,
		"remove a network interface by name.")
	c.Flags().StringVarP(
		&file, "file", "f", "",
		`will deploy the capsule spec at the given path. Cannot be used together with any of the other configuration flags.
The spec is the Platform Capsule spec defined at https://docs.rig.dev/api/platformv1#capsule`,
	)
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

func (c *Cmd) imageCompletions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	if capsule.CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var imageIDs []string

	if c.Scope.GetCurrentContext() == nil || c.Scope.GetCurrentContext().GetAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Image().List(ctx, connect.NewRequest(
		&image.ListRequest{
			CapsuleId: capsule.CapsuleID,
			ProjectId: c.Scope.GetCurrentContext().GetProject(),
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
