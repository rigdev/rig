package plugin

import (
	"context"
	"slices"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
)

type Plugin interface {
	Run(context.Context, pipeline.CapsuleRequest) error
	Stop(context.Context)
}

type Step struct {
	step   v1alpha1.Step
	logger logr.Logger
	plugin Plugin
}

func NewStep(step v1alpha1.Step, logger logr.Logger) (*Step, error) {
	var p Plugin
	var err error
	switch step.Plugin {
	default:
		p, err = NewExternalPlugin(step.Plugin, logger, step.Config)
	}
	if err != nil {
		return nil, err
	}

	return &Step{
		step:   step,
		logger: logger,
		plugin: p,
	}, nil
}

func (s *Step) Apply(ctx context.Context, req pipeline.CapsuleRequest) error {
	if len(s.step.Namespaces) > 0 {
		if !slices.Contains(s.step.Namespaces, req.Capsule().Namespace) {
			return nil
		}
	}

	s.logger.Info("running plugin", "plugin", s.step.Plugin)

	return s.plugin.Run(ctx, req)
}

func (s *Step) Stop(ctx context.Context) {
	s.plugin.Stop(ctx)
}
