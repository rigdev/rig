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
	steps  []Step[CapsuleRequest]
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

func (p *CapsulePipeline) AddStep(step Step[CapsuleRequest]) {
	p.steps = append(p.steps, step)
}

func (p *CapsulePipeline) RunCapsule(
	ctx context.Context,
	capsule *v1alpha2.Capsule,
	opts ...CapsuleRequestOption,
) (*Result, error) {
	req := newCapsuleRequest(p, capsule, opts...)
	return ExecuteRequest(ctx, req, p.steps, true)
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

// TODO This ExecuteableRequest type construction is a bit messy
// Find a better abstraction
type ExecutableRequest[T Request] interface {
	GetRequest() T
	GetBase() *RequestBase
}

func ExecuteRequest[T Request](
	ctx context.Context,
	req ExecutableRequest[T],
	steps []Step[T],
	commit bool,
) (*Result, error) {
	result, err := executeRequestInner(ctx, req, steps, commit)
	if errors.IsFailedPrecondition(err) {
		return nil, err
	} else if err != nil {
		if !req.GetBase().dryRun {
			if err := req.GetBase().Strategies.UpdateStatusWithError(ctx, err); err != nil {
				return nil, err
			}
		}

		return nil, err
	}

	return result, nil
}

func executeRequestInner[T Request](
	ctx context.Context,
	req ExecutableRequest[T], steps []Step[T],
	commit bool,
) (*Result, error) {
	if err := req.GetBase().Strategies.LoadExistingObjects(ctx); err != nil {
		return nil, err
	}

	for {
		result := req.GetBase().PrepareRequest()
		req.GetBase().logger.Info("run steps", "existing_objects", maps.Keys(req.GetBase().existingObjects))

		for _, s := range steps {
			if err := s.Apply(ctx, req.GetRequest()); err != nil {
				return nil, err
			}
		}

		if !commit {
			return result, nil
		}

		changes, err := req.GetBase().Commit(ctx)
		if errors.IsAborted(err) {
			req.GetBase().logger.Info("retry running steps", "error", errors.MessageOf(err))
			continue
		} else if err != nil {
			req.GetBase().logger.Error(err, "error committing changes")
			return nil, err
		}

		for key, c := range changes {
			obj := req.GetBase().newObjects[key].Materialized
			if obj == nil {
				obj = req.GetBase().newObjects[key].New
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
