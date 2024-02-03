package plugin

import (
	"context"
	"slices"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/errors"
)

type Plugin interface {
	Run(context.Context, pipeline.Request) error
}

var _pluginFactories = map[string]func(config map[string]string) (Plugin, error){
	"template":      NewTemplatePlugin,
	"sidecar":       NewSidecarPlugin,
	"initContainer": NewInitContainerPlugin,
}

type Step struct {
	step v1alpha1.Step
}

func NewStep(step v1alpha1.Step) *Step {
	return &Step{
		step: step,
	}
}

func (s *Step) Apply(ctx context.Context, req pipeline.Request) error {
	pf, ok := _pluginFactories[s.step.Plugin]
	if !ok {
		return errors.InvalidArgumentErrorf("unknown plugin '%v'", s.step.Plugin)
	}

	if len(s.step.Namespaces) > 0 {
		if !slices.Contains(s.step.Namespaces, req.Capsule().Namespace) {
			return nil
		}
	}

	p, err := pf(s.step.Config)
	if err != nil {
		return err
	}

	return p.Run(ctx, req)
}
