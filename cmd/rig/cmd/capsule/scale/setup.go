package scale

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	disable             bool
	overwriteAutoscaler bool
	forceDeploy         bool
)

var (
	autoscalerPath string
)

var vflags VerticalFlags

var (
	replicas              uint32
	utilizationPercentage uint32
	minReplicas           uint32
	maxReplicas           uint32
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Scope    scope.Scope
	Prompter common.Prompter
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	scale := &cobra.Command{
		Use:               "scale",
		Short:             "Scale and inspect the resources of the capsule",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		GroupID:           capsule.DeploymentGroupID,
	}

	scaleGet := &cobra.Command{
		Use:   "get [capsule]",
		Short: "Displays the resources (container size) and replicas of the capsule",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.get),
	}
	scale.AddCommand(scaleGet)

	scaleVertical := &cobra.Command{
		Use:   "vertical [capsule]",
		Short: "Vertically scaling the capsule (setting the container size)",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.vertical),
	}
	scaleVertical.Flags().StringVar(&vflags.CPURequest, "request-cpu", "", "Minimum CPU cores per container")
	scaleVertical.Flags().StringVar(&vflags.MemoryRequest, "request-memory", "", "Minimum memory per container")
	scaleVertical.Flags().StringVar(&vflags.CPULimit, "limit-cpu", "", "Maximum CPU cores per container")
	scaleVertical.Flags().StringVar(&vflags.MemoryLimit, "limit-memory", "", "Maximum memory per container")
	scaleVertical.Flags().Uint32Var(&vflags.GPULimit, "limit-gpu", 0, "Maximum number of GPUs per container")
	scaleVertical.Flags().StringVar(&vflags.GPUType, "gpu-type", "", "GPU type")
	scaleVertical.MarkFlagsRequiredTogether("limit-gpu", "gpu-type")

	scaleVertical.Flags().BoolVarP(
		&forceDeploy,
		"force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes",
	)
	scale.AddCommand(scaleVertical)

	scaleHorizontal := &cobra.Command{
		Use:   "horizontal [capsule]",
		Short: "Horizontally scaling the capsule (setting the number of replicas and configuring the autoscaler)",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.horizontal),
	}

	scaleHorizontal.Flags().Uint32VarP(&replicas, "replicas", "r", 0, "number of replicas to scale to")
	scaleHorizontal.Flags().BoolVarP(
		&overwriteAutoscaler, "overwrite-autoscaler", "a", false, "if the autoscaler is enabled, this flag is "+
			"necessary to set the replicas. It will disable the autoscaler.",
	)
	scaleHorizontal.Flags().BoolVarP(
		&forceDeploy,
		"force-deploy", "f", false, "Abort the current rollout if one is in progress and deploy the changes",
	)
	scale.AddCommand(scaleHorizontal)

	scaleHorizontalAuto := &cobra.Command{
		Use:   "autoscale [capsule]",
		Short: "Configure the autoscaler for horizontal scaling",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.autoscale),
	}

	scaleHorizontal.SetCompletionCommandGroupID("horizontal")

	scaleHorizontalAuto.Flags().Uint32VarP(
		&utilizationPercentage,
		"utilization-percentage", "u", 0, "CPU utilization percentage for the autoscaler. 1 <= 100",
	)
	scaleHorizontalAuto.Flags().Uint32Var(&minReplicas, "min-replicas", 0, "minimum replicas")
	scaleHorizontalAuto.Flags().Uint32Var(&maxReplicas, "max-replicas", 0, "maximum replicas")
	scaleHorizontalAuto.Flags().BoolVarP(
		&disable,
		"disable", "d", false, "Disable the autoscaler, fixing the capsule with its current number of replicas",
	)
	scaleHorizontalAuto.Flags().StringVar(
		&autoscalerPath,
		"path", "", `If given, reads the configuration for the autoscaler from the file. Accepts json or yaml.
If other flags are given as well, they overwrite their fields in the configuration at 'path'.`,
	)
	scaleHorizontal.AddCommand(scaleHorizontalAuto)

	parent.AddCommand(scale)
}

func (c *Cmd) completions(
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

type VerticalFlags struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
	GPUType       string
	GPULimit      uint32
}

func (v VerticalFlags) Empty(cmd *cobra.Command) bool {
	return (v.CPURequest == "" &&
		!cmd.Flags().Changed("cpu-limit") &&
		v.MemoryRequest == "" &&
		!cmd.Flags().Changed("memory-limit") &&
		v.GPUType == "" &&
		!cmd.Flags().Changed("gpu-limit"))
}
