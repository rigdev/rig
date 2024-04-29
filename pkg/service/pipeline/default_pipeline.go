package pipeline

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/service/capabilities"
	"k8s.io/apimachinery/pkg/runtime"
)

func (s *service) initializePipeline(ctx context.Context) error {
	p, err := CreateDefaultPipeline(ctx, s.client.Scheme(), s.capSvc, s.cfg, s.pluginManager, s.logger)
	if err != nil {
		return err
	}

	s.pipeline = p
	return nil
}

func CreateDefaultPipeline(
	ctx context.Context,
	scheme *runtime.Scheme,
	capSvc capabilities.Service,
	cfg *v1alpha1.OperatorConfig,
	pluginManager *plugin.Manager,
	logger logr.Logger,
) (*pipeline.CapsulePipeline, error) {
	steps, err := GetDefaultPipelineSteps(ctx, capSvc, cfg, pluginManager, logger)
	if err != nil {
		return nil, err
	}

	pipeline := pipeline.NewCapsulePipeline(cfg, scheme, logger)
	for _, step := range steps {
		pipeline.AddStep(step)
	}

	for _, step := range cfg.Pipeline.Steps {
		ps, err := pluginManager.NewStep(step, logger)
		if err != nil {
			return nil, err
		}

		pipeline.AddStep(ps)
	}

	return pipeline, nil
}

func GetDefaultPipelineSteps(
	ctx context.Context,
	capSvc capabilities.Service,
	cfg *v1alpha1.OperatorConfig,
	pluginManager *plugin.Manager,
	logger logr.Logger,
) ([]pipeline.Step[pipeline.CapsuleRequest], error) {
	capabilities, err := capSvc.Get(ctx)
	if err != nil {
		return nil, err
	}

	var steps []pipeline.Step[pipeline.CapsuleRequest]

	steps = append(steps,
		NewServiceAccountStep(),
		NewDeploymentStep(),
		NewVPAStep(cfg),
		NewNetworkStep(cfg),
	)

	if cfg.Pipeline.RoutesStep.Plugin != "" {
		routesStep, err := NewRoutesStep(cfg, pluginManager, logger)
		if err != nil {
			return nil, err
		}

		steps = append(steps, routesStep)
	}

	steps = append(steps,
		NewCronJobStep(),
	)

	if capabilities.GetHasPrometheusServiceMonitor() {
		steps = append(steps, NewServiceMonitorStep(cfg))
	}

	return steps, nil
}

func NewRoutesStep(cfg *v1alpha1.OperatorConfig,
	pluginManager *plugin.Manager,
	logger logr.Logger,
) (pipeline.Step[pipeline.CapsuleRequest], error) {
	routesStep, err := pluginManager.NewStep(v1alpha1.Step{
		EnableForPlatform: true,
		Plugins: []v1alpha1.Plugin{
			{
				Name:   cfg.Pipeline.RoutesStep.Plugin,
				Config: cfg.Pipeline.RoutesStep.Config,
			},
		},
	}, logger)
	if err != nil {
		return nil, err
	}

	return routesStep, nil
}
