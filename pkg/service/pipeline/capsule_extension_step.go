package pipeline

import (
	"context"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
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

func (s *CapsuleExtensionStep) WatchObjectStatus(ctx context.Context, namespace, capsule string, callback pipeline.ObjectStatusCallback) error {
	// TODO: We want to opt out if not relevant for this capsule.
	return s.step.WatchObjectStatus(ctx, namespace, capsule, callback)
}

func (s *CapsuleExtensionStep) PluginIDs() []uuid.UUID {
	return s.step.PluginIDs()
}

type CapsuleExtensionValidationStep struct {
	cfg *v1alpha1.OperatorConfig
}

func NewCapsuleExtensionValidationStep(cfg *v1alpha1.OperatorConfig) *CapsuleExtensionValidationStep {
	return &CapsuleExtensionValidationStep{
		cfg: cfg,
	}
}

func (s *CapsuleExtensionValidationStep) Apply(ctx context.Context, req pipeline.CapsuleRequest, opts pipeline.Options) error {
	for name := range req.Capsule().Spec.Extensions {
		if _, ok := s.cfg.Pipeline.CapsuleExtensions[name]; !ok {
			return errors.UnimplementedErrorf("capsule extension '%s' not supported by cluster", name)
		}
	}

	return nil
}

func (s *CapsuleExtensionValidationStep) WatchObjectStatus(ctx context.Context, namespace, capsule string, callback pipeline.ObjectStatusCallback) error {
	return nil
}

func (s *CapsuleExtensionValidationStep) PluginIDs() []uuid.UUID {
	return nil
}
