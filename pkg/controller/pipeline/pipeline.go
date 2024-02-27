package pipeline

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/errors"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LabelOwnedByCapsule = "rig.dev/owned-by-capsule"
)

type Object struct {
	Current      client.Object
	New          client.Object
	Materialized client.Object
}

type objectKey struct {
	client.ObjectKey
	schema.GroupVersionKind
}

func (ok objectKey) String() string {
	return fmt.Sprintf("%s/%s", ok.GroupVersionKind.String(), ok.ObjectKey.String())
}

func (ok objectKey) MarshalLog() interface{} {
	return struct {
		Group     string `json:"group,omitempty"`
		Version   string `json:"version,omitempty"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Name:      ok.Name,
		Namespace: ok.Namespace,
		Group:     ok.Group,
		Kind:      ok.Kind,
		Version:   ok.Version,
	}
}

type Pipeline struct {
	client client.Client
	config *configv1alpha1.OperatorConfig
	scheme *runtime.Scheme
	// TODO Use zap instead
	logger logr.Logger
	steps  []Step
}

func New(
	cc client.Client,
	config *configv1alpha1.OperatorConfig,
	scheme *runtime.Scheme,
	logger logr.Logger,
) *Pipeline {
	p := &Pipeline{
		client: cc,
		config: config,
		scheme: scheme,
		logger: logger,
	}

	return p
}

func (p *Pipeline) AddStep(step Step) {
	p.steps = append(p.steps, step)
}

func (p *Pipeline) RunCapsule(ctx context.Context, capsule *v1alpha2.Capsule, opts ...CapsuleRequestOption) (*Result, error) {
	req := newCapsuleRequest(p, capsule, opts...)

	result, err := p.runSteps(ctx, req)
	if errors.IsFailedPrecondition(err) {
		return nil, err
	} else if err != nil {
		if !req.dryRun {
			if err := req.updateStatusError(ctx, err); err != nil {
				return nil, err
			}
		}

		return nil, err
	}

	return result, nil
}

type OutputObject struct {
	ObjectKey objectKey
	Object    client.Object
	State     ResourceState
}
type Result struct {
	InputObjects  []client.Object
	OutputObjects []OutputObject
	Objects       []*Object
}

func (p *Pipeline) runSteps(ctx context.Context, req *capsuleRequest) (*Result, error) {
	if err := req.loadExisting(ctx); err != nil {
		return nil, err
	}

	for {
		result := req.prepare()

		p.logger.Info("run steps", "current_objects", maps.Keys(req.currentObjects))

		for _, s := range p.steps {
			if err := s.Apply(ctx, req); err != nil {
				return nil, err
			}
		}

		changes, err := req.commit(ctx)
		if errors.IsAborted(err) {
			p.logger.Error(err, "retry running steps")
			continue
		} else if err != nil {
			p.logger.Error(err, "error committing changes")
			return nil, err
		}

		for key, c := range changes {
			obj := req.objects[key].Materialized
			if obj == nil {
				obj = req.objects[key].New
			}
			result.OutputObjects = append(result.OutputObjects, OutputObject{
				ObjectKey: key,
				Object:    obj,
				State:     c.state,
			})
		}
		return result, nil
	}
}
