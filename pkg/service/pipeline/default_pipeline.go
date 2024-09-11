package pipeline

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/rigdev/rig/plugins/capsulesteps/cron_jobs"
	"github.com/rigdev/rig/plugins/capsulesteps/deployment"
	"github.com/rigdev/rig/plugins/capsulesteps/service_account"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	_serviceAccountPlatformConfig = `
useExisting: true
`
)

func (s *service) initializePipeline() error {
	execCtx := plugin.NewExecutionContext(context.Background())

	p, err := CreateDefaultPipeline(execCtx, s.client.Scheme(), s.vm, s.cfg, s.pluginManager, s.logger)
	if err != nil {
		execCtx.Stop()
		return err
	}

	s.pipeline = p
	s.execCtx = execCtx
	return nil
}

func CreateDefaultPipeline(
	execCtx plugin.ExecutionContext,
	scheme *runtime.Scheme,
	vm scheme.VersionMapper,
	cfg *v1alpha1.OperatorConfig,
	pluginManager *plugin.Manager,
	logger logr.Logger,
) (*pipeline.CapsulePipeline, error) {
	steps, err := GetDefaultPipelineSteps(execCtx, cfg, pluginManager, logger)
	if err != nil {
		return nil, err
	}

	pipeline := pipeline.NewCapsulePipeline(cfg, scheme, vm, logger)
	for _, step := range steps {
		pipeline.AddStep(step)
	}

	for _, step := range cfg.Pipeline.Steps {
		ps, err := pluginManager.NewStep(execCtx, step, logger)
		if err != nil {
			return nil, err
		}

		pipeline.AddStep(ps)
	}

	return pipeline, nil
}

func GetDefaultPipelineSteps(
	execCtx plugin.ExecutionContext,
	cfg *v1alpha1.OperatorConfig,
	pluginManager *plugin.Manager,
	logger logr.Logger,
) ([]pipeline.Step[pipeline.CapsuleRequest], error) {
	deploymentPlugin := deployment.Name
	if cfg.Pipeline.DeploymentStep.Plugin != "" {
		deploymentPlugin = cfg.Pipeline.DeploymentStep.Plugin
	}
	deploymentStep, err := NewCapsulePluginStep(execCtx, deploymentPlugin,
		cfg.Pipeline.DeploymentStep.Config, pluginManager, logger, true)
	if err != nil {
		return nil, err
	}

	serviceAccountPlugin := service_account.Name
	if cfg.Pipeline.ServiceAccountStep.Plugin != "" {
		serviceAccountPlugin = cfg.Pipeline.ServiceAccountStep.Plugin
	}
	serviceAccountStep, err := NewCapsulePluginStep(execCtx, serviceAccountPlugin,
		cfg.Pipeline.ServiceAccountStep.Config, pluginManager, logger, false)
	if err != nil {
		return nil, err
	}

	serviceAccountPlatformStep, err := NewRigPlatformCapsulePluginStep(
		execCtx, service_account.Name, _serviceAccountPlatformConfig, pluginManager, logger)
	if err != nil {
		return nil, err
	}

	steps := []pipeline.Step[pipeline.CapsuleRequest]{
		deploymentStep,
		serviceAccountStep,
		serviceAccountPlatformStep,
	}

	if cfg.Pipeline.VPAStep.Plugin != "" {
		vpaStep, err := NewCapsulePluginStep(execCtx, cfg.Pipeline.VPAStep.Plugin,
			cfg.Pipeline.VPAStep.Config, pluginManager, logger, true)
		if err != nil {
			return nil, err
		}

		steps = append(steps, vpaStep)
	}

	if cfg.Pipeline.RoutesStep.Plugin != "" {
		routesStep, err := NewCapsulePluginStep(execCtx, cfg.Pipeline.RoutesStep.Plugin,
			cfg.Pipeline.RoutesStep.Config, pluginManager, logger, true)
		if err != nil {
			return nil, err
		}

		steps = append(steps, routesStep)
	}

	cronJobsPlugin := cron_jobs.Name
	if cfg.Pipeline.CronJobsStep.Plugin != "" {
		cronJobsPlugin = cfg.Pipeline.CronJobsStep.Plugin
	}
	cronjobStep, err := NewCapsulePluginStep(execCtx, cronJobsPlugin,
		cfg.Pipeline.CronJobsStep.Config, pluginManager, logger, true)
	if err != nil {
		return nil, err
	}
	steps = append(steps,
		cronjobStep,
	)

	if cfg.Pipeline.ServiceMonitorStep.Plugin != "" {
		serviceMonitorStep, err := NewCapsulePluginStep(execCtx, cfg.Pipeline.ServiceMonitorStep.Plugin,
			cfg.Pipeline.ServiceMonitorStep.Config, pluginManager, logger, true)
		if err != nil {
			return nil, err
		}

		steps = append(steps, serviceMonitorStep)
	}

	steps = append(steps, NewCapsuleExtensionValidationStep(cfg))
	for name, capsuleStep := range cfg.Pipeline.CapsuleExtensions {
		if capsuleStep.Plugin != "" {
			step, err := NewCapsulePluginStep(execCtx, capsuleStep.Plugin, capsuleStep.Config, pluginManager, logger, false)
			if err != nil {
				return nil, err
			}

			steps = append(steps, NewCapsuleExtensionStep(name, step))
		}
	}

	return steps, nil
}

func NewCapsulePluginStep(
	execCtx plugin.ExecutionContext,
	pluginName, pluginConfig string,
	pluginManager *plugin.Manager,
	logger logr.Logger,
	enableForPlatform bool,
) (pipeline.Step[pipeline.CapsuleRequest], error) {
	pluginStep, err := pluginManager.NewStep(
		execCtx,
		v1alpha1.Step{
			EnableForPlatform: enableForPlatform,
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

func NewRigPlatformCapsulePluginStep(
	execCtx plugin.ExecutionContext,
	pluginName string,
	pluginConfig string,
	pluginManager *plugin.Manager,
	logger logr.Logger,
) (pipeline.Step[pipeline.CapsuleRequest], error) {
	pluginStep, err := pluginManager.NewStep(
		execCtx,
		v1alpha1.Step{
			Match: v1alpha1.CapsuleMatch{
				Namespaces:        []string{"rig-system"},
				Names:             []string{"rig-platform"},
				EnableForPlatform: true,
			},
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
