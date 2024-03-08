package plugin

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gobwas/glob"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type Plugin interface {
	Run(context.Context, pipeline.CapsuleRequest) error
	Stop(context.Context)
}

type Step struct {
	step    v1alpha1.Step
	logger  logr.Logger
	plugins []Plugin
	matcher Matcher
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
	c := req.Capsule()
	if !s.matcher.Match(c.Namespace, c.Name, c.Annotations) {
		return nil
	}
	for i, p := range s.plugins {
		tag := s.step.Tag
		if s.step.Plugins[i].Tag != "" {
			tag = s.step.Plugins[i].Tag
		}
		s.logger.Info(
			"running plugin",
			"plugin", s.step.Plugins[i].Name, "capsule_id", c.Name, "namespace", c.Namespace, "tag", tag,
		)
		if err := p.Run(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

func (s *Step) Stop(ctx context.Context) {
	for _, p := range s.plugins {
		p.Stop(ctx)
	}
	s.plugins = nil
}

type Matcher struct {
	namespaces []glob.Glob
	capsules   []glob.Glob
	selector   labels.Selector
}

func NewMatcher(namespaces, capsules []string, selector metav1.LabelSelector) (Matcher, error) {
	s, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return Matcher{}, err
	}

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
		selector:   s,
	}, nil
}

func (m Matcher) Match(namespace, capsule string, capsuleLabels map[string]string) bool {
	if !m.selector.Matches(labels.Set(capsuleLabels)) {
		return false
	}
	if !match(m.namespaces, namespace) {
		return false
	}
	if !match(m.capsules, capsule) {
		return false
	}
	return true
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
