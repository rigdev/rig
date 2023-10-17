package scale

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	outputJSON          bool
	disable             bool
	overwriteAutoscaler bool
)

var (
	requestCPU    string
	requestMemory string
	limitCPU      string
	limitMemory   string
)

var (
	replicas              uint32
	utilizationPercentage uint32
	minReplicas           uint32
	maxReplicas           uint32
)

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
}

func (r Cmd) Setup(parent *cobra.Command) {
	scale := &cobra.Command{
		Use:   "scale",
		Short: "Scale and inspect the resources of the capsule",
	}

	scaleGet := &cobra.Command{
		Use:               "get",
		Short:             "Displays the resources (container size) and replicas of the capsule",
		Args:              cobra.NoArgs,
		RunE:              r.get,
		ValidArgsFunction: common.NoCompletions,
	}
	scaleGet.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	scaleGet.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	scale.AddCommand(scaleGet)

	scaleVertical := &cobra.Command{
		Use:               "vertical",
		Short:             "Vertically scaling the capsule (setting the container size)",
		Args:              cobra.NoArgs,
		RunE:              r.vertical,
		ValidArgsFunction: common.NoCompletions,
	}
	scaleVertical.Flags().StringVar(&requestCPU, "request-cpu", "", "Minimum CPU cores per container")
	scaleVertical.Flags().StringVar(&requestMemory, "request-memory", "", "Minimum memory per container")
	scaleVertical.RegisterFlagCompletionFunc("request-cpu", common.NoCompletions)
	scaleVertical.RegisterFlagCompletionFunc("request-memory", common.NoCompletions)

	scaleVertical.Flags().StringVar(&limitCPU, "limit-cpu", "", "Maximum CPU cores per container")
	scaleVertical.Flags().StringVar(&limitMemory, "limit-memory", "", "Maximum memory per container")
	scaleVertical.RegisterFlagCompletionFunc("limit-cpu", common.NoCompletions)
	scaleVertical.RegisterFlagCompletionFunc("limit-memory", common.NoCompletions)
	scale.AddCommand(scaleVertical)

	scaleHorizontal := &cobra.Command{
		Use:               "horizontal",
		Short:             "Horizontally scaling the capsule (setting the number of replicas and configuring the autoscaler)",
		Args:              cobra.NoArgs,
		RunE:              r.horizontal,
		ValidArgsFunction: common.NoCompletions,
	}
	scaleHorizontal.Flags().Uint32VarP(&replicas, "replicas", "r", 0, "number of replicas to scale to")
	scaleHorizontal.Flags().BoolVarP(&overwriteAutoscaler, "overwrite-autoscaler", "o", false, "if the autoscaler is enabled, this flag is necessary to set the replicas. It will disable the autoscaler.")
	scaleHorizontal.RegisterFlagCompletionFunc("replicas", common.NoCompletions)
	scaleHorizontal.RegisterFlagCompletionFunc("overwrite-autoscaler", common.NoCompletions)
	scale.AddCommand(scaleHorizontal)

	scaleHorizontalAuto := &cobra.Command{
		Use:               "autoscale",
		Short:             "Configure the autoscaler for horizontal scaling",
		Args:              cobra.NoArgs,
		RunE:              r.autoscale,
		ValidArgsFunction: common.NoCompletions,
	}
	scaleHorizontalAuto.Flags().Uint32VarP(&utilizationPercentage, "utilization-percentage", "u", 0, "CPU utilization percentage for the autoscaler. 1 <= 100")
	scaleHorizontalAuto.Flags().Uint32Var(&minReplicas, "min-replicas", 0, "minimum replicas")
	scaleHorizontalAuto.Flags().Uint32Var(&maxReplicas, "max-replicas", 0, "maximum replicas")
	scaleHorizontalAuto.Flags().BoolVarP(&disable, "disable", "d", false, "Disable the autoscaler, fixing the capsule with its current number of replicas")
	scaleHorizontalAuto.RegisterFlagCompletionFunc("min-replicas", common.NoCompletions)
	scaleHorizontalAuto.RegisterFlagCompletionFunc("max-replicas", common.NoCompletions)
	scaleHorizontalAuto.RegisterFlagCompletionFunc("disable", common.NoCompletions)
	scaleHorizontal.AddCommand(scaleHorizontalAuto)

	parent.AddCommand(scale)
}
