package pipeline

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/controller/mod"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service interface {
	DryRun(ctx context.Context,
		cfg *v1alpha1.OperatorConfig,
		namespace, capsuleName string,
		spec *v1alpha2.Capsule,
		opts ...pipeline.CapsuleRequestOption) (*pipeline.Result, error)
}

type ModUsed struct {
	Namespace string
	Capsule   string
	Mod       string
}

func NewService(
	cfg config.Service,
	client client.Client,
	capSvc capabilities.Service,
	logger logr.Logger,
	modManager *mod.Manager,
) Service {
	return &service{
		cfg:        cfg,
		client:     client,
		capSvc:     capSvc,
		logger:     logger,
		modManager: modManager,
	}
}

type service struct {
	cfg        config.Service
	client     client.Client
	capSvc     capabilities.Service
	logger     logr.Logger
	modManager *mod.Manager
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
		cfg = s.cfg.Operator()
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

	steps, err := controller.GetDefaultPipelineSteps(ctx, s.capSvc, cfg, s.modManager, s.logger)
	if err != nil {
		return nil, err
	}

	p := pipeline.NewCapsulePipeline(s.client, s.client, cfg, scheme.New(), s.logger)
	for _, step := range steps {
		p.AddStep(step)
	}

	for _, step := range cfg.Pipeline.Steps {
		ps, err := s.modManager.NewStep(step, s.logger)
		if err != nil {
			return nil, err
		}

		p.AddStep(ps)
		defer ps.Stop(ctx)
	}

	return p.RunCapsule(ctx, spec, append(opts, pipeline.WithDryRun())...)
}
