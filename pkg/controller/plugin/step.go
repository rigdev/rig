package plugin

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/gobwas/glob"
	"github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type Step struct {
	step    v1alpha1.Step
	logger  logr.Logger
	plugins []*pluginExecutor
	matcher Matcher
	name    string
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

func (s *Step) Apply(ctx context.Context, req pipeline.CapsuleRequest, opts pipeline.Options) error {
	c := req.Capsule()
	if !s.matcher.Match(c.Namespace, c.Name, c.Annotations) {
		return nil
	}
	for i, p := range s.plugins {
		logger := s.logger.WithValues(
			"step", s.name, "plugin_idx", i+1, "plugin", s.step.Plugins[i].GetPlugin(),
			"capsule_id", c.Name, "namespace", c.Namespace,
		)
		logger.Info("running plugin")
		if err := p.Run(ctx, req, opts); err != nil {
			logger.Error(err, "plugin failed")
			return fmt.Errorf("plugin #%v (%s) failed: %w", i+1, p.name, err)
		}
	}
	return nil
}

func (s *Step) WatchObjectStatus(
	ctx context.Context,
	capsule *v1alpha2.Capsule,
	callback pipeline.ObjectStatusCallback,
) error {
	// TODO: We need annotations here.
	if !s.matcher.Match(capsule.GetNamespace(), capsule.GetName(), capsule.GetAnnotations()) {
		for _, p := range s.plugins {
			callback.UpdateStatus(capsule.GetNamespace(), capsule.GetName(), p.id, &plugin.ObjectStatusChange{
				Change: &plugin.ObjectStatusChange_Checkpoint_{},
			})
		}
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	for _, p := range s.plugins {
		wg.Add(1)
		go func(p *pluginExecutor) {
			defer wg.Done()
			err := p.WatchObjectStatus(ctx, capsule, callback)
			if errors.IsUnimplemented(err) {
				callback.UpdateStatus(capsule.GetNamespace(), capsule.GetName(), p.id, &plugin.ObjectStatusChange{
					Change: &plugin.ObjectStatusChange_Checkpoint_{},
				})
			} else if err != nil {
				if !errors.IsCanceled(err) {
					s.logger.Error(err, "error getting status")
				}

				select {
				default:
				case errChan <- err:
				}
			}
		}(p)
	}

	go func() {
		wg.Wait()
		select {
		default:
		case errChan <- nil:
		}
	}()

	return <-errChan
}

func (s *Step) PluginIDs() []uuid.UUID {
	var plugins []uuid.UUID
	for _, p := range s.plugins {
		plugins = append(plugins, p.id)
	}
	return plugins
}

func (s *Step) ComputeConfig(ctx context.Context, req pipeline.CapsuleRequest) pipeline.StepConfigResult {
	res := pipeline.StepConfigResult{
		Name: s.Name(),
	}
	for _, p := range s.plugins {
		config, err := p.ComputeConfig(ctx, req)
		if err != nil {
			res.Plugins = append(res.Plugins, pipeline.PluginConfig{
				Name: p.name,
				Err:  err.Error(),
			})
		} else {
			res.Plugins = append(res.Plugins, pipeline.PluginConfig{
				Name:   p.name,
				Config: config,
			})
		}
	}
	return res
}

func (s *Step) Name() string {
	return s.name
}

type Matcher struct {
	namespaces        []glob.Glob
	capsules          []glob.Glob
	selector          labels.Selector
	enableForPlatform bool
}

func NewMatcher(match v1alpha1.CapsuleMatch) (Matcher, error) {
	s, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: match.Annotations,
	})
	if err != nil {
		return Matcher{}, err
	}

	nsGlobs, err := makeGlobs(match.Namespaces)
	if err != nil {
		return Matcher{}, err
	}
	cGlobs, err := makeGlobs(match.Names)
	if err != nil {
		return Matcher{}, err
	}
	return Matcher{
		namespaces:        nsGlobs,
		capsules:          cGlobs,
		selector:          s,
		enableForPlatform: match.EnableForPlatform,
	}, nil
}

func (m Matcher) Match(namespace, capsule string, capsuleAnnotations map[string]string) bool {
	if !m.selector.Matches(labels.Set(capsuleAnnotations)) {
		return false
	}
	if !match(m.namespaces, namespace) {
		return false
	}
	if !match(m.capsules, capsule) {
		return false
	}
	if capsule == "rig-platform" && !m.enableForPlatform {
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
