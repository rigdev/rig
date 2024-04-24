package pipeline

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/plugins/cron_jobs"
	"github.com/rigdev/rig/plugins/deployment"
	"github.com/rigdev/rig/plugins/service_account"
	"k8s.io/apimachinery/pkg/runtime"
)

func (s *service) initializePipeline(ctx context.Context) error {
	p, err := CreateDefaultPipeline(ctx, s.client.Scheme(), s.cfg, s.pluginManager, s.logger)
	if err != nil {
		return err
	}

	s.pipeline = p
	return nil
}

func CreateDefaultPipeline(
	ctx context.Context,
	scheme *runtime.Scheme,
	cfg *v1alpha1.OperatorConfig,
	pluginManager *plugin.Manager,
	logger logr.Logger,
) (*pipeline.CapsulePipeline, error) {
	steps, err := GetDefaultPipelineSteps(ctx, cfg, pluginManager, logger)
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
	_ context.Context,
	cfg *v1alpha1.OperatorConfig,
	pluginManager *plugin.Manager,
	logger logr.Logger,
) ([]pipeline.Step[pipeline.CapsuleRequest], error) {
	serviceAccountPlugin := service_account.Name
	if cfg.Pipeline.ServiceAccountStep.Plugin != "" {
		serviceAccountPlugin = cfg.Pipeline.ServiceAccountStep.Plugin
	}
	serviceAccountStep, err := NewCapsulePluginStep(serviceAccountPlugin,
		cfg.Pipeline.ServiceAccountStep.Config, pluginManager, logger)
	if err != nil {
		return nil, err
	}

	deploymentPlugin := deployment.Name
	if cfg.Pipeline.DeploymentStep.Plugin != "" {
		deploymentPlugin = cfg.Pipeline.DeploymentStep.Plugin
	}
	deploymentStep, err := NewCapsulePluginStep(deploymentPlugin,
		cfg.Pipeline.DeploymentStep.Config, pluginManager, logger)
	if err != nil {
		return nil, err
	}

	steps := []pipeline.Step[pipeline.CapsuleRequest]{
		serviceAccountStep,
		deploymentStep,
	}

	if cfg.Pipeline.VPAStep.Plugin != "" {
		vpaStep, err := NewCapsulePluginStep(cfg.Pipeline.VPAStep.Plugin,
			cfg.Pipeline.VPAStep.Config, pluginManager, logger)
		if err != nil {
			return nil, err
		}

		steps = append(steps, vpaStep)
	}

	if cfg.Pipeline.RoutesStep.Plugin != "" {
		routesStep, err := NewCapsulePluginStep(cfg.Pipeline.RoutesStep.Plugin,
			cfg.Pipeline.RoutesStep.Config, pluginManager, logger)
		if err != nil {
			return nil, err
		}

		steps = append(steps, routesStep)
	}

	cronJobsPlugin := cron_jobs.Name
	if cfg.Pipeline.CronJobsStep.Plugin != "" {
		cronJobsPlugin = cfg.Pipeline.CronJobsStep.Plugin
	}
	cronjobStep, err := NewCapsulePluginStep(cronJobsPlugin,
		cfg.Pipeline.CronJobsStep.Config, pluginManager, logger)
	if err != nil {
		return nil, err
	}
	steps = append(steps,
		cronjobStep,
	)

	if cfg.Pipeline.ServiceMonitorStep.Plugin != "" {
		serviceMonitorStep, err := NewCapsulePluginStep(cfg.Pipeline.ServiceMonitorStep.Plugin,
			cfg.Pipeline.ServiceMonitorStep.Config, pluginManager, logger)
		if err != nil {
			return nil, err
		}

		steps = append(steps, serviceMonitorStep)
	}

	return steps, nil
}

func NewCapsulePluginStep(
	pluginName, pluginConfig string,
	pluginManager *plugin.Manager,
	logger logr.Logger,
) (pipeline.Step[pipeline.CapsuleRequest], error) {
	pluginStep, err := pluginManager.NewStep(v1alpha1.Step{
		EnableForPlatform: true,
		Plugins: []v1alpha1.Plugin{
			{
				Name:   pluginName,
				Config: pluginConfig,
			},
		},
	}, logger)
	if err != nil {
		return nil, err
	}

	return pluginStep, nil
}
