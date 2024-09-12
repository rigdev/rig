package pipeline

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/rigdev/rig/pkg/service/capabilities"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service interface {
	GetDefaultPipeline() *pipeline.CapsulePipeline
	DryRun(ctx context.Context,
		cfg *v1alpha1.OperatorConfig,
		namespace, capsuleName string,
		spec *v1alpha2.Capsule,
		opts ...pipeline.CapsuleRequestOption) (*pipeline.Result, error)
	DryRunPluginConfig(ctx context.Context,
		cfg *v1alpha1.OperatorConfig,
		namespace, capsuleName string,
		capsuleSpec *v1alpha2.Capsule,
	) (pipeline.PluginConfigResult, error)
}

type PluginUsed struct {
	Namespace string
	Capsule   string
	Plugin    string
}

func NewService(
	cfg *v1alpha1.OperatorConfig,
	client client.Client,
	capSvc capabilities.Service,
	logger logr.Logger,
	pluginManager *plugin.Manager,
	lc fx.Lifecycle,
	sh fx.Shutdowner,
) Service {
	s := &service{
		cfg:           cfg,
		client:        client,
		capSvc:        capSvc,
		logger:        logger,
		pluginManager: pluginManager,
		vm:            scheme.NewVersionMapper(client),
	}

	lc.Append(fx.StartHook(func() error {
		if err := s.initializePipeline(); err != nil {
			return err
		}

		if sh != nil {
			go func() {
				<-s.execCtx.Context().Done()
				s.logger.Info("default pipeline plugins terminated, restarting")
				_ = sh.Shutdown(fx.ExitCode(1))
			}()
		}

		return nil
	}))

	return s
}

type service struct {
	cfg           *v1alpha1.OperatorConfig
	client        client.Client
	capSvc        capabilities.Service
	logger        logr.Logger
	pluginManager *plugin.Manager
	pipeline      *pipeline.CapsulePipeline
	execCtx       plugin.ExecutionContext
	vm            scheme.VersionMapper
}

func (s *service) GetDefaultPipeline() *pipeline.CapsulePipeline {
	return s.pipeline
}

// DryRun implements Service.
func (s *service) DryRun(
	ctx context.Context,
	cfg *v1alpha1.OperatorConfig,
	namespace, capsuleName string,
	capsuleSpec *v1alpha2.Capsule,
	opts ...pipeline.CapsuleRequestOption,
) (*pipeline.Result, error) {
	p, err := s.setupDryRunPipeline(ctx, cfg, namespace, capsuleName, capsuleSpec)
	if err != nil {
		return nil, err
	}
	return p.RunCapsule(ctx, capsuleSpec, s.client, append(opts, pipeline.WithDryRun())...)
}

func (s *service) DryRunPluginConfig(ctx context.Context,
	cfg *v1alpha1.OperatorConfig,
	namespace, capsuleName string,
	capsuleSpec *v1alpha2.Capsule,
) (pipeline.PluginConfigResult, error) {
	p, err := s.setupDryRunPipeline(ctx, cfg, namespace, capsuleName, capsuleSpec)
	if err != nil {
		return pipeline.PluginConfigResult{}, err
	}

	return p.ComputeConfig(ctx, capsuleSpec, s.client)
}

func (s *service) setupDryRunPipeline(
	ctx context.Context, cfg *v1alpha1.OperatorConfig, namespace, capsuleName string, capsuleSpec *v1alpha2.Capsule,
) (*pipeline.CapsulePipeline, error) {
	if capsuleSpec == nil {
		capsuleSpec = &v1alpha2.Capsule{}
		if err := errors.FromK8sClient(s.client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      capsuleName,
		}, capsuleSpec)); err != nil {
			return nil, err
		}
	} else {
		// Load existing status object.
		currentSpec := &v1alpha2.Capsule{}
		if err := errors.FromK8sClient(s.client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      capsuleName,
		}, currentSpec)); errors.IsNotFound(err) {
			// Noop.
		} else if err != nil {
			return nil, err
		} else {
			capsuleSpec.Status = currentSpec.Status
			capsuleSpec.UID = currentSpec.GetUID()
		}
	}

	if len(capsuleSpec.GetUID()) == 0 {
		capsuleSpec.SetUID(types.UID("dry-run-spec"))
	}

	var p *pipeline.CapsulePipeline
	if cfg == nil {
		p = s.GetDefaultPipeline()
	} else {
		execCtx := plugin.NewExecutionContext(ctx)

		steps, err := GetDefaultPipelineSteps(execCtx, cfg, s.pluginManager, s.logger)
		if err != nil {
			return nil, err
		}

		p = pipeline.NewCapsulePipeline(cfg, scheme.New(), s.vm, s.logger)
		for _, step := range steps {
			p.AddStep(step)
		}

		for idx, step := range cfg.Pipeline.Steps {
			ps, err := s.pluginManager.NewStep(execCtx, step, s.logger, customStepName(idx))
			if err != nil {
				return nil, err
			}

			p.AddStep(ps)
		}
	}
	return p, nil
}

func customStepName(idx int) string {
	return fmt.Sprintf("customStep#%v", idx)
}
