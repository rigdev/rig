package plugin

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gobwas/glob"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
)

type Plugin interface {
	Run(context.Context, pipeline.CapsuleRequest) error
	Stop(context.Context)
}

type Step struct {
	step    v1alpha1.Step
	logger  logr.Logger
	plugin  Plugin
	matcher Matcher
}

func NewStep(step v1alpha1.Step, logger logr.Logger) (*Step, error) {
	p, err := NewExternalPlugin(step.Plugin, logger, step.Config)
	if err != nil {
		return nil, err
	}

	matcher, err := NewMatcher(step.Namespaces, step.Capsules)
	if err != nil {
		return nil, err
	}

	return &Step{
		step:    step,
		logger:  logger,
		plugin:  p,
		matcher: matcher,
	}, nil
}

func makeGlobs(strings []string) ([]glob.Glob, error) {
	var res []glob.Glob
	for _, s := range strings {
		g, err := glob.Compile(s)
		if err != nil {
			return nil, err
		}
		res = append(res, g)
	}
	return res, nil
}

func (s *Step) Apply(ctx context.Context, req pipeline.CapsuleRequest) error {
	cap := req.Capsule()
	if !s.matcher.Match(cap.Name, cap.Namespace) {
		return nil
	}
	s.logger.Info("running plugin", "plugin", s.step.Plugin)
	return s.plugin.Run(ctx, req)
}

func (s *Step) Stop(ctx context.Context) {
	s.plugin.Stop(ctx)
}

type Matcher struct {
	namespaces []glob.Glob
	capsules   []glob.Glob
}

func NewMatcher(namespaces, capsules []string) (Matcher, error) {
	nsGlobs, err := makeGlobs(namespaces)
	if err != nil {
		return Matcher{}, err
	}
	cGlobs, err := makeGlobs(capsules)
	if err != nil {
		return Matcher{}, err
	}
	return Matcher{
		namespaces: nsGlobs,
		capsules:   cGlobs,
	}, nil
}

func (m Matcher) Match(namespace, capsule string) bool {
	return match(m.namespaces, namespace) && match(m.capsules, capsule)
}

func match(globs []glob.Glob, pattern string) bool {
	if len(globs) == 0 {
		return true
	}
	for _, g := range globs {
		if g.Match(pattern) {
			return true
		}
	}
	return false
}
