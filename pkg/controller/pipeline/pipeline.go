package pipeline

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/errors"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Object struct {
	Current      client.Object
	New          client.Object
	Materialized client.Object
}

type ObjectKey struct {
	client.ObjectKey
	schema.GroupVersionKind
}

func sortedKeys(keys []ObjectKey) []ObjectKey {
	slices.SortStableFunc(keys, func(k1, k2 ObjectKey) int { return strings.Compare(k1.String(), k2.String()) })
	return keys
}

func (ok ObjectKey) String() string {
	return fmt.Sprintf("%s/%s", ok.GroupVersionKind.String(), ok.ObjectKey.String())
}

func (ok ObjectKey) MarshalLog() interface{} {
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

type CapsulePipeline struct {
	client client.Client
	reader client.Reader
	config *configv1alpha1.OperatorConfig
	scheme *runtime.Scheme
	// TODO Use zap instead
	logger logr.Logger
	steps  []CapsuleStep
}

func NewCapsulePipeline(
	cc client.Client,
	cr client.Reader,
	config *configv1alpha1.OperatorConfig,
	scheme *runtime.Scheme,
	logger logr.Logger,
) *CapsulePipeline {
	p := &CapsulePipeline{
		client: cc,
		reader: cr,
		config: config,
		scheme: scheme,
		logger: logger,
	}

	return p
}

func (p *CapsulePipeline) AddStep(step CapsuleStep) {
	p.steps = append(p.steps, step)
}

func (p *CapsulePipeline) RunCapsule(
	ctx context.Context,
	capsule *v1alpha2.Capsule,
	opts ...CapsuleRequestOption,
) (*Result, error) {
	req := newCapsuleRequest(p, capsule, opts...)
	var steps []func(context.Context, CapsuleRequest) error
	for _, s := range p.steps {
		steps = append(steps, s.Apply)
	}
	return ExecuteRequest(ctx, CapsuleRequest(req), &req.RequestBase, steps, true)
}

type OutputObject struct {
	ObjectKey ObjectKey
	Object    client.Object
	State     ResourceState
}

type Result struct {
	InputObjects  []client.Object
	OutputObjects []OutputObject
}

// TODO Fix this req/base shit
func ExecuteRequest[T Request](
	ctx context.Context,
	req T, base *RequestBase,
	steps []func(context.Context, T) error,
	commit bool,
) (*Result, error) {
	result, err := executeRequestInner(ctx, req, base, steps, commit)
	if errors.IsFailedPrecondition(err) {
		return nil, err
	} else if err != nil {
		if !base.dryRun {
			if err := base.Strategies.UpdateStatusWithError(ctx, err); err != nil {
				return nil, err
			}
		}

		return nil, err
	}

	return result, nil
}

func executeRequestInner[T Request](
	ctx context.Context,
	req T, base *RequestBase,
	steps []func(context.Context, T) error, commit bool,
) (*Result, error) {
	if err := base.Strategies.LoadExistingObjects(ctx); err != nil {
		return nil, err
	}

	for {
		result := base.PrepareRequest()

		base.logger.Info("run steps", "existing_objects", maps.Keys(base.existingObjects))

		for _, s := range steps {
			if err := s(ctx, req); err != nil {
				return nil, err
			}
		}

		if !commit {
			return result, nil
		}

		changes, err := base.Commit(ctx)
		if errors.IsAborted(err) {
			base.logger.Error(err, "retry running steps")
			continue
		} else if err != nil {
			base.logger.Error(err, "error committing changes")
			return nil, err
		}

		for key, c := range changes {
			obj := base.newObjects[key].Materialized
			if obj == nil {
				obj = base.newObjects[key].New
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
