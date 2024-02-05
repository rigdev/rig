package plugin

import (
	"context"
	"reflect"
	"slices"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/errors"
)

type Plugin interface {
	Run(context.Context, pipeline.Request) error
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
	if len(s.step.Namespaces) > 0 {
		if !slices.Contains(s.step.Namespaces, req.Capsule().Namespace) {
			return nil
		}
	}

	raw, err := s.step.Plugin.GetPlugin()
	if err != nil {
		return err
	}

	var p Plugin
	switch v := raw.(type) {
	case *v1alpha1.ObjectPlugin:
		p = NewObjectPlugin(v)
	case *v1alpha1.SidecarPlugin:
		p = NewSidecarPlugin(v)
	case *v1alpha1.InitContainerPlugin:
		p = NewInitContainerPlugin(v)
	default:
		return errors.InvalidArgumentErrorf("unknown plugin '%v'", reflect.TypeOf(v))
	}

	return p.Run(ctx, req)
}
