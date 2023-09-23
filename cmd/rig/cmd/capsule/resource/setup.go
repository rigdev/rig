package resource

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	outputJSON bool
)

var (
	requestCPU    string
	requestMemory string
	limitCPU      string
	limitMemory   string
)

var (
	replicas int
)

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
}

func (r Cmd) Setup(parent *cobra.Command) {
	resource := &cobra.Command{
		Use:   "resource",
		Short: "Scale and inspect the resources of the capsule",
	}

	resourcesGet := &cobra.Command{
		Use:               "get",
		Short:             "Displays the resources (container size) and replicas of the capsule",
		Args:              cobra.NoArgs,
		RunE:              r.get,
		ValidArgsFunction: common.NoCompletions,
	}
	resourcesGet.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	resourcesGet.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	resource.AddCommand(resourcesGet)

	resourcesScale := &cobra.Command{
		Use:               "scale",
		Short:             "Sets the container resources for the capsule",
		Args:              cobra.NoArgs,
		RunE:              r.scale,
		ValidArgsFunction: common.NoCompletions,
	}
	resourcesScale.Flags().StringVar(&requestCPU, "request-cpu", "", "Minimum CPU cores per container")
	resourcesScale.Flags().StringVar(&requestMemory, "request-memory", "", "Minimum memory per container")
	resourcesScale.RegisterFlagCompletionFunc("request-cpu", common.NoCompletions)
	resourcesScale.RegisterFlagCompletionFunc("request-memory", common.NoCompletions)

	resourcesScale.Flags().StringVar(&limitCPU, "limit-cpu", "", "Maximum CPU cores per container")
	resourcesScale.Flags().StringVar(&limitMemory, "limit-memory", "", "Maximum memory per container")
	resourcesScale.RegisterFlagCompletionFunc("limit-cpu", common.NoCompletions)
	resourcesScale.RegisterFlagCompletionFunc("limit-memory", common.NoCompletions)

	resourcesScale.Flags().IntVarP(&replicas, "replicas", "r", -1, "number of replicas to scale to")
	resourcesScale.RegisterFlagCompletionFunc("replicas", common.NoCompletions)
	resource.AddCommand(resourcesScale)

	parent.AddCommand(resource)
}
