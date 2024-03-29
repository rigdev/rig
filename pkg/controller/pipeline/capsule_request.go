package pipeline

import (
	"context"
	"fmt"

	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/roclient"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LabelOwnedByCapsule = "rig.dev/owned-by-capsule"
)

// CapsuleRequest contains a single reconcile request for a given capsule.
// It contains both the set of existing kubernetes objects owned by the capsule
// and the set of objects recorded to be applied after all steps in the pipeline has been executed (called 'new' objects).
// The set of existing objects cannot be modified (as the interface does not allow for writing to Kubernetes)
// but there are both read and write access to the set of new objects.
//
//nolint:lll
type CapsuleRequest interface {
	Request
	// Capsule returns a deepcopy of the capsule object being reconciled.
	Capsule() *v1alpha2.Capsule
	// MarkUsedObject marks the object as used by the Capsule which will be present in the Capsule's Status
	MarkUsedObject(res v1alpha2.UsedResource) error
}

type capsuleRequest struct {
	RequestBase
	capsule       *v1alpha2.Capsule
	usedResources []v1alpha2.UsedResource
}

type CapsuleRequestOption interface {
	apply(*capsuleRequest)
}

type withDryRun struct{}

func (withDryRun) apply(r *capsuleRequest) { r.dryRun = true }
func WithDryRun() CapsuleRequestOption {
	return withDryRun{}
}

type withAdditionalResources struct {
	resources []*pipeline.Object
}

func (w withAdditionalResources) apply(r *capsuleRequest) {
	reader := roclient.NewReader(r.scheme)
	for _, o := range w.resources {
		proposal, err := obj.DecodeAny([]byte(o.Content), r.scheme)
		if err != nil {
			continue
		}

		if err := reader.AddObject(proposal); err != nil {
			continue
		}
	}

	r.reader = roclient.NewLayeredReader(r.reader, reader)
}

func WithAdditionalResources(resources []*pipeline.Object) CapsuleRequestOption {
	return withAdditionalResources{resources}
}

type withForce struct{}

func (withForce) apply(r *capsuleRequest) { r.force = true }
func WithForce() CapsuleRequestOption {
	return withForce{}
}

func NewCapsuleRequest(p *CapsulePipeline, capsule *v1alpha2.Capsule, opts ...CapsuleRequestOption) CapsuleRequest {
	return newCapsuleRequest(p, capsule, opts...)
}

func newCapsuleRequest(p *CapsulePipeline, capsule *v1alpha2.Capsule, opts ...CapsuleRequestOption) *capsuleRequest {
	r := &capsuleRequest{
		RequestBase: NewRequestBase(p.client, p.reader, p.config, p.scheme, p.logger, nil, capsule),
		capsule:     capsule,
	}
	// TODO This is an ugly hack. Find a better solution
	// Good rule of thumb: If the Rust compiler would throw a fit, do it differently.
	r.Strategies = r

	for _, opt := range opts {
		opt.apply(r)
	}

	if capsule.Status != nil {
		r.observedGeneration = capsule.Status.ObservedGeneration
	}

	r.logger.Info("created capsule request",
		"generation", capsule.Generation,
		"observed_generation", r.observedGeneration,
		"resource_version", capsule.ResourceVersion,
		"dry_run", r.dryRun,
		"force", r.force,
	)

	return r
}

func (r *capsuleRequest) Capsule() *v1alpha2.Capsule {
	return r.capsule.DeepCopy()
}

func (r *capsuleRequest) getGVK(obj client.Object) (schema.GroupVersionKind, error) {
	gvks, _, err := r.scheme.ObjectKinds(obj)
	if err != nil {
		r.logger.Error(err, "invalid object type")
		return schema.GroupVersionKind{}, err
	}

	return gvks[0], nil
}

func (r *capsuleRequest) GetKey(obj client.Object) (ObjectKey, error) {
	if obj.GetName() == "" {
		obj.SetName(r.capsule.Name)
	}
	obj.SetNamespace(r.capsule.Namespace)

	gvk, err := r.getGVK(obj)
	if err != nil {
		return ObjectKey{}, err
	}

	return r.namedObjectKey(obj.GetName(), gvk), nil
}

func (r *capsuleRequest) namedObjectKey(name string, gvk schema.GroupVersionKind) ObjectKey {
	return ObjectKey{
		ObjectKey: types.NamespacedName{
			Name:      name,
			Namespace: r.capsule.Namespace,
		},
		GroupVersionKind: gvk,
	}
}

func (r *capsuleRequest) MarkUsedObject(res v1alpha2.UsedResource) error {
	r.usedResources = append(r.usedResources, res)
	return nil
}

func (r *capsuleRequest) LoadExistingObjects(ctx context.Context) error {
	// Read all status objects.
	s := r.capsule.Status
	if s == nil {
		return nil
	}

	for _, o := range s.OwnedResources {
		if o.Ref == nil {
			continue
		}
		gk := schema.GroupKind{
			Kind: o.Ref.Kind,
		}
		if o.Ref.APIGroup != nil {
			gk.Group = *o.Ref.APIGroup
		}

		gvk, err := LookupGVK(gk)
		if err != nil {
			return err
		}

		ro, err := r.scheme.New(gvk)
		if err != nil {
			return err
		}

		co, ok := ro.(client.Object)
		if !ok {
			continue
		}

		co.SetName(o.Ref.Name)
		co.SetNamespace(r.capsule.Namespace)
		co.GetObjectKind().SetGroupVersionKind(gvk)
		if err := r.reader.Get(ctx, client.ObjectKeyFromObject(co), co); kerrors.IsNotFound(err) {
			// Okay it doesn't exist, ignore the resource.
			continue
		} else if err != nil {
			return err
		}

		r.existingObjects[r.namedObjectKey(o.Ref.Name, gvk)] = co
	}

	return nil
}

func (r *capsuleRequest) Prepare() {
	r.usedResources = nil
}

func (r *capsuleRequest) UpdateStatusWithChanges(
	ctx context.Context,
	changes map[ObjectKey]*Change,
) error {
	capsule, ok := r.requestObject.(*v1alpha2.Capsule)
	if !ok {
		return fmt.Errorf("object given to capsuleStatusUpdater had wrong type: %T", r.requestObject)
	}
	capsuleCopy := capsule.DeepCopy()

	status := &v1alpha2.CapsuleStatus{
		ObservedGeneration: r.observedGeneration,
	}

	for _, key := range sortedKeys(maps.Keys(changes)) {
		key := key
		change := changes[key]
		or := v1alpha2.OwnedResource{
			Ref: &v1.TypedLocalObjectReference{
				APIGroup: &key.Group,
				Kind:     key.Kind,
				Name:     key.Name,
			},
			State: string(change.state),
		}
		switch change.state {
		case ResourceStateCreated, ResourceStateUpdated, ResourceStateDeleted:
			if !change.applied {
				or.State = string(ResourceStateChangePending)
			}
		}
		if change.err != nil {
			or.Message = change.err.Error()
		}
		status.OwnedResources = append(status.OwnedResources, or)
	}

	status.UsedResources = r.usedResources

	capsuleCopy.Status = status

	if err := r.client.Status().Update(ctx, capsuleCopy); err != nil {
		return err
	}

	capsule.Status = status
	capsule.SetResourceVersion(capsule.GetResourceVersion())

	return nil
}

func (r *capsuleRequest) UpdateStatusWithError(ctx context.Context, err error) error {
	capsule, ok := r.requestObject.(*v1alpha2.Capsule)
	if !ok {
		return fmt.Errorf("object had unexpected type %T", r.requestObject)
	}
	capsuleCopy := capsule.DeepCopy()

	status := &v1alpha2.CapsuleStatus{
		ObservedGeneration: r.observedGeneration,
		Errors:             []string{err.Error()},
	}

	if capsuleCopy.Status != nil {
		status.OwnedResources = capsuleCopy.Status.OwnedResources
		status.UsedResources = capsuleCopy.Status.UsedResources
	}

	capsuleCopy.Status = status

	if err := r.client.Status().Update(ctx, capsuleCopy); err != nil {
		return err
	}

	capsule.Status = status
	capsule.SetResourceVersion(capsuleCopy.GetResourceVersion())

	return nil
}

func (*capsuleRequest) OwnedLabel() string {
	return LabelOwnedByCapsule
}
