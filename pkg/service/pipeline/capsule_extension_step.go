package pipeline

import (
	"context"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/uuid"
)

type CapsuleExtensionStep struct {
	name string
	step pipeline.Step[pipeline.CapsuleRequest]
}

func NewCapsuleExtensionStep(name string, step pipeline.Step[pipeline.CapsuleRequest]) *CapsuleExtensionStep {
	return &CapsuleExtensionStep{
		name: name,
		step: step,
	}
}

func (s *CapsuleExtensionStep) Apply(ctx context.Context, req pipeline.CapsuleRequest, opts pipeline.Options) error {
	if _, ok := req.Capsule().Spec.Extensions[s.name]; ok {
		return s.step.Apply(ctx, req, opts)
	}

	return nil
}

func (s *CapsuleExtensionStep) WatchObjectStatus(
	ctx context.Context, capsule *v1alpha2.Capsule, callback pipeline.ObjectStatusCallback,
) error {
	if _, ok := capsule.Spec.Extensions[s.name]; !ok {
		return nil
	}
	return s.step.WatchObjectStatus(ctx, capsule, callback)
}

func (s *CapsuleExtensionStep) PluginIDs() []uuid.UUID {
	return s.step.PluginIDs()
}

func (s *CapsuleExtensionStep) ComputeConfig(
	ctx context.Context, req pipeline.CapsuleRequest,
) pipeline.StepConfigResult {
	return s.step.ComputeConfig(ctx, req)
}

func (s *CapsuleExtensionStep) Name() string {
	return s.name
}

type CapsuleExtensionValidationStep struct {
	cfg *v1alpha1.OperatorConfig
}

func NewCapsuleExtensionValidationStep(cfg *v1alpha1.OperatorConfig) *CapsuleExtensionValidationStep {
	return &CapsuleExtensionValidationStep{
		cfg: cfg,
	}
}

func (s *CapsuleExtensionValidationStep) Apply(
	_ context.Context, req pipeline.CapsuleRequest, _ pipeline.Options,
) error {
	for name := range req.Capsule().Spec.Extensions {
		if _, ok := s.cfg.Pipeline.CapsuleExtensions[name]; !ok {
			return errors.UnimplementedErrorf("capsule extension '%s' not supported by cluster", name)
		}
	}

	return nil
}

func (s *CapsuleExtensionValidationStep) WatchObjectStatus(
	_ context.Context, _ *v1alpha2.Capsule, _ pipeline.ObjectStatusCallback,
) error {
	return nil
}

func (s *CapsuleExtensionValidationStep) PluginIDs() []uuid.UUID {
	return nil
}

func (s *CapsuleExtensionValidationStep) Name() string {
	return "extension_validation"
}

func (s *CapsuleExtensionValidationStep) ComputeConfig(
	_ context.Context, _ pipeline.CapsuleRequest,
) pipeline.StepConfigResult {
	return pipeline.StepConfigResult{}
}
