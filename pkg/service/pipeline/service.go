package pipeline

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/plugin"
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
) Service {
	s := &service{
		cfg:           cfg,
		client:        client,
		capSvc:        capSvc,
		logger:        logger,
		pluginManager: pluginManager,
	}
	lc.Append(fx.StartHook(s.initializePipeline))
	return s
}

type service struct {
	cfg           *v1alpha1.OperatorConfig
	client        client.Client
	capSvc        capabilities.Service
	logger        logr.Logger
	pluginManager *plugin.Manager
	pipeline      *pipeline.CapsulePipeline
}

func (s *service) GetDefaultPipeline() *pipeline.CapsulePipeline {
	return s.pipeline
}

// DryRun implements Service.
func (s *service) DryRun(
	ctx context.Context,
	cfg *v1alpha1.OperatorConfig,
	namespace, capsuleName string,
	spec *v1alpha2.Capsule,
	opts ...pipeline.CapsuleRequestOption,
) (*pipeline.Result, error) {
	if cfg == nil {
		cfg = s.cfg
	}
	if spec == nil {
		spec = &v1alpha2.Capsule{}
		if err := s.client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      capsuleName,
		}, spec); err != nil {
			return nil, err
		}
	}

	if len(spec.GetUID()) == 0 {
		spec.SetUID(types.UID("dry-run-spec"))
	}

	steps, err := GetDefaultPipelineSteps(ctx, cfg, s.pluginManager, s.logger)
	if err != nil {
		return nil, err
	}

	p := pipeline.NewCapsulePipeline(cfg, scheme.New(), s.logger)
	for _, step := range steps {
		p.AddStep(step)
	}

	for _, step := range cfg.Pipeline.Steps {
		ps, err := s.pluginManager.NewStep(step, s.logger)
		if err != nil {
			return nil, err
		}

		p.AddStep(ps)
		defer ps.Stop(ctx)
	}

	return p.RunCapsule(ctx, spec, s.client, append(opts, pipeline.WithDryRun())...)
}
