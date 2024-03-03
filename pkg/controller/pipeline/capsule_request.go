package pipeline

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/errors"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type CapsuleRequest interface {
	Config() *configv1alpha1.OperatorConfig
	Scheme() *runtime.Scheme
	Client() client.Client
	Capsule() *v1alpha2.Capsule
	GetCurrent(obj client.Object) error
	GetNew(obj client.Object) error
	Set(obj client.Object) error
	Delete(obj client.Object) error
	MarkUsedResource(res v1alpha2.UsedResource)
}

type capsuleRequest struct {
	pipeline           *Pipeline
	capsule            *v1alpha2.Capsule
	currentObjects     map[objectKey]client.Object
	objects            map[objectKey]*Object
	observedGeneration int64
	usedResources      []v1alpha2.UsedResource
	logger             logr.Logger
	dryRun             bool
	force              bool
}

type CapsuleRequestOption interface {
	apply(*capsuleRequest)
}

type withDryRun struct{}

func (withDryRun) apply(r *capsuleRequest) { r.dryRun = true }
func WithDryRun() CapsuleRequestOption {
	return withDryRun{}
}

type withForce struct{}

func (withForce) apply(r *capsuleRequest) { r.force = true }
func WithForce() CapsuleRequestOption {
	return withForce{}
}

func NewCapsuleRequest(p *Pipeline, capsule *v1alpha2.Capsule, opts ...CapsuleRequestOption) CapsuleRequest {
	return newCapsuleRequest(p, capsule, opts...)
}

func newCapsuleRequest(p *Pipeline, capsule *v1alpha2.Capsule, opts ...CapsuleRequestOption) *capsuleRequest {
	r := &capsuleRequest{
		pipeline:       p,
		capsule:        capsule,
		currentObjects: map[objectKey]client.Object{},
		objects:        map[objectKey]*Object{},
		usedResources:  []v1alpha2.UsedResource{},
		logger:         p.logger.WithValues("capsule", capsule.Name),
	}

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

func (r *capsuleRequest) Config() *configv1alpha1.OperatorConfig {
	return r.pipeline.config.DeepCopy()
}

func (r *capsuleRequest) Scheme() *runtime.Scheme {
	return r.pipeline.scheme
}

func (r *capsuleRequest) Capsule() *v1alpha2.Capsule {
	return r.capsule.DeepCopy()
}

func (r *capsuleRequest) Client() client.Client {
	return r.pipeline.client
}

func (r *capsuleRequest) GetCurrent(obj client.Object) error {
	key, err := r.getKey(obj)
	if err != nil {
		return err
	}

	o, ok := r.objects[key]
	if !ok {
		return errors.NotFoundErrorf("object '%v' of type '%v' not found", key.Name, key.GroupVersionKind)
	}

	if o.Current == nil {
		return errors.NotFoundErrorf("object '%v' of type '%v' has no existing version", key.Name, key.GroupVersionKind)
	}

	return r.pipeline.scheme.Converter().Convert(o.Current, obj, nil)
}

func (r *capsuleRequest) GetNew(obj client.Object) error {
	key, err := r.getKey(obj)
	if err != nil {
		return err
	}

	o, ok := r.objects[key]
	if !ok {
		return errors.NotFoundErrorf("object '%v' of type '%v' not found", key.Name, key.GroupVersionKind)
	}

	if o.New == nil {
		return errors.NotFoundErrorf("object '%v' of type '%v' has no new version", key.Name, key.GroupVersionKind)
	}

	return r.pipeline.scheme.Converter().Convert(o.New, obj, nil)
}

func (r *capsuleRequest) Set(obj client.Object) error {
	key, err := r.getKey(obj)
	if err != nil {
		return err
	}

	o, ok := r.objects[key]
	if !ok {
		o = &Object{}
	}
	o.New = obj
	r.objects[key] = o
	return nil
}

func (r *capsuleRequest) Delete(obj client.Object) error {
	key, err := r.getKey(obj)
	if err != nil {
		return err
	}

	o, ok := r.objects[key]
	if ok {
		o.New = nil
	}

	return nil
}

func (r *capsuleRequest) getGVK(obj client.Object) (schema.GroupVersionKind, error) {
	gvks, _, err := r.pipeline.scheme.ObjectKinds(obj)
	if err != nil {
		r.logger.Error(err, "invalid object type")
		return schema.GroupVersionKind{}, err
	}

	return gvks[0], nil
}

func (r *capsuleRequest) getKey(obj client.Object) (objectKey, error) {
	if obj.GetName() == "" {
		obj.SetName(r.capsule.Name)
	}
	obj.SetNamespace(r.capsule.Namespace)

	gvk, err := r.getGVK(obj)
	if err != nil {
		return objectKey{}, err
	}

	obj.SetNamespace(r.capsule.Namespace)
	return r.namedObjectKey(obj.GetName(), gvk), nil
}

func (r *capsuleRequest) namedObjectKey(name string, gvk schema.GroupVersionKind) objectKey {
	return objectKey{
		ObjectKey: types.NamespacedName{
			Name:      name,
			Namespace: r.capsule.Namespace,
		},
		GroupVersionKind: gvk,
	}
}

func (r *capsuleRequest) MarkUsedResource(res v1alpha2.UsedResource) {
	r.usedResources = append(r.usedResources, res)
}

func (r *capsuleRequest) loadExisting(ctx context.Context) error {
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

		ro, err := r.pipeline.scheme.New(gvk)
		if err != nil {
			return err
		}

		co, ok := ro.(client.Object)
		if !ok {
			continue
		}

		co.SetName(o.Ref.Name)
		co.SetNamespace(r.capsule.Namespace)
		if err := r.pipeline.client.Get(ctx, client.ObjectKeyFromObject(co), co); kerrors.IsNotFound(err) {
			// Okay it doesn't exist, ignore the resource.
			continue
		} else if err != nil {
			return err
		}

		r.currentObjects[r.namedObjectKey(o.Ref.Name, gvk)] = co
	}

	return nil
}

func (r *capsuleRequest) prepare() *Result {
	result := &Result{}
	r.usedResources = nil
	r.objects = map[objectKey]*Object{}
	for _, k := range sortedKeys(maps.Keys(r.currentObjects)) {
		o := r.currentObjects[k]
		r.objects[k] = &Object{
			Current: o.DeepCopyObject().(client.Object),
		}
		result.InputObjects = append(result.InputObjects, o.DeepCopyObject().(client.Object))
	}
	return result
}

type change struct {
	state   ResourceState
	applied bool
	err     error
}

func (r *capsuleRequest) commit(ctx context.Context) (map[objectKey]*change, error) {
	allKeys := sortedKeys(maps.Keys(r.objects))

	// Prepare all the new objects with default labels / owner refs.
	for _, key := range allKeys {
		cObj := r.objects[key]
		if cObj.New == nil {
			continue
		}

		cObj.New.GetObjectKind().SetGroupVersionKind(key.GroupVersionKind)
		normalizeObject(key, cObj.New)

		labels := cObj.New.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[LabelOwnedByCapsule] = r.capsule.Name
		cObj.New.SetLabels(labels)

		if err := controllerutil.SetControllerReference(r.capsule, cObj.New, r.pipeline.scheme); err != nil {
			return nil, err
		}
	}

	changes := map[objectKey]*change{}

	// Dry run to detect no-op vs create vs update.
	for _, key := range allKeys {
		cObj := r.objects[key]

		if cObj.Current == nil {
			materializedObj := cObj.New.DeepCopyObject().(client.Object)
			if err := r.pipeline.client.Create(ctx, materializedObj, client.DryRunAll); kerrors.IsConflict(err) {
				return nil, errors.FailedPreconditionErrorf("new object version available for '%v'", key)
			} else if kerrors.IsAlreadyExists(err) || kerrors.IsInvalid(err) {
				o, newErr := r.pipeline.scheme.New(key.GroupVersionKind)
				if newErr != nil {
					return nil, err
				}

				co := o.(client.Object)
				if getErr := r.pipeline.client.Get(ctx, key.ObjectKey, co); kerrors.IsNotFound(getErr) {
					r.logger.Info("configuration is invalid", "object", key, "error", err)
					return nil, err
				} else if getErr != nil {
					return nil, fmt.Errorf("could not get existing object: %w", getErr)
				}

				if r.force || IsOwnedBy(r.capsule, co) {
					r.logger.Info("object exists but not in status, retrying", "object", key)
					r.currentObjects[key] = normalizeObject(key, co)
					return nil, errors.AbortedErrorf("object exists but not in capsule status")
				}

				r.logger.Info("create object skipped, not owned by controller", "object", key)
				changes[key] = &change{state: ResourceStateAlreadyExists}
				continue
			} else if err != nil {
				return nil, fmt.Errorf("could not render create to %s: %w", key.GroupVersionKind, err)
			}

			cObj.Materialized = normalizeObject(key, materializedObj)

			r.logger.Info("create object", "object", key)
			changes[key] = &change{state: ResourceStateCreated}
			continue
		}

		if !(r.force || IsOwnedBy(r.capsule, cObj.Current)) {
			r.logger.Info("update object skipped, not owned by controller", "object", key)
			changes[key] = &change{state: ResourceStateAlreadyExists}
			continue
		}

		if cObj.New == nil {
			r.logger.Info("delete object", "object", key)
			changes[key] = &change{state: ResourceStateDeleted}
			continue
		}

		materializedObj := cObj.New.DeepCopyObject().(client.Object)
		materializedObj.GetObjectKind().SetGroupVersionKind(cObj.Current.GetObjectKind().GroupVersionKind())

		// Dry run to fully materialize the new spec.
		materializedObj.SetResourceVersion(cObj.Current.GetResourceVersion())
		if r.force && r.dryRun {
			// TODO: If just force, we probably need to delete and re-create. Let's explore workarounds.
			materializedObj.SetOwnerReferences(cObj.Current.GetOwnerReferences())
		}
		r.logger.Info("generating materialized version", "object", key)
		if err := r.pipeline.client.Update(ctx, materializedObj, client.DryRunAll); kerrors.IsConflict(err) {
			return nil, errors.FailedPreconditionErrorf("new object version available for '%v'", key)
		} else if err != nil {
			return nil, fmt.Errorf("could not render update to %s: %w", key.GroupVersionKind, err)
		}

		if ObjectsEquals(cObj.Current, materializedObj) {
			r.logger.Info("update object skipped, not changed", "object", key)
			changes[key] = &change{state: ResourceStateUnchanged}
			continue
		}

		cObj.Materialized = normalizeObject(key, materializedObj)

		r.logger.Info("update object", "object", key)
		changes[key] = &change{state: ResourceStateUpdated}
	}

	// Skip update if no changes.
	if r.observedGeneration == r.capsule.Generation {
		r.logger.Info("already at generation", "generation", r.observedGeneration)
		hasChanges := false
		for _, change := range changes {
			switch change.state {
			case ResourceStateUpdated, ResourceStateCreated, ResourceStateDeleted:
				hasChanges = true
			}
		}
		if !hasChanges {
			r.logger.Info("no changes to apply", "generation", r.observedGeneration)
			return changes, nil
		}
	}

	if r.dryRun {
		return changes, nil
	}

	if err := r.updateStatusChanges(ctx, changes, r.observedGeneration); err != nil {
		return nil, err
	}

	var errs []error
	for _, key := range sortedKeys(maps.Keys(changes)) {
		change := changes[key]
		if err := r.applyChange(ctx, key, change.state); err != nil {
			change.err = err
			errs = append(errs, err)
		} else {
			change.applied = true
		}
	}

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	if err := r.updateStatusChanges(ctx, changes, r.capsule.Generation); err != nil {
		return nil, err
	}

	return changes, nil
}

func (r *capsuleRequest) applyChange(ctx context.Context, key objectKey, state ResourceState) error {
	switch state {
	case ResourceStateUpdated:
		r.logger.Info("update object", "object", key)
		obj := r.objects[key]
		obj.New.SetResourceVersion(obj.Current.GetResourceVersion())
		if err := r.pipeline.client.Update(ctx, obj.New); err != nil {
			return fmt.Errorf("could not update %s: %w", key.GroupVersionKind, err)
		}

	case ResourceStateCreated:
		r.logger.Info("create object", "object", key)
		obj := r.objects[key]
		if err := r.pipeline.client.Create(ctx, obj.New); err != nil {
			return fmt.Errorf("could not create %s: %w", key.GroupVersionKind, err)
		}

	case ResourceStateDeleted:
		r.logger.Info("delete object", "object", key)
		obj := r.objects[key]
		if err := r.pipeline.client.Delete(ctx, obj.Current); err != nil {
			return fmt.Errorf("could not update %s: %w", key.GroupVersionKind, err)
		}
	}

	return nil
}

type ResourceState string

const (
	ResourceStateDeleted       ResourceState = "deleted"
	ResourceStateUpdated       ResourceState = "updated"
	ResourceStateUnchanged     ResourceState = "unchanged"
	ResourceStateCreated       ResourceState = "created"
	ResourceStateFailed        ResourceState = "failed"
	ResourceStateAlreadyExists ResourceState = "alreadyExists"
	ResourceStateChangePending ResourceState = "changePending"
)

func (r *capsuleRequest) updateStatusChanges(
	ctx context.Context,
	changes map[objectKey]*change,
	generation int64,
) error {
	capsule := r.capsule.DeepCopy()

	status := &v1alpha2.CapsuleStatus{
		ObservedGeneration: generation,
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

	capsule.Status = status

	if err := r.pipeline.client.Status().Update(ctx, capsule); err != nil {
		return err
	}

	r.observedGeneration = generation
	r.capsule.Status = status
	r.capsule.SetResourceVersion(capsule.GetResourceVersion())

	return nil
}

func (r *capsuleRequest) updateStatusError(ctx context.Context, err error) error {
	capsule := r.capsule.DeepCopy()

	status := &v1alpha2.CapsuleStatus{
		ObservedGeneration: r.observedGeneration,
		Errors:             []string{err.Error()},
	}

	if capsule.Status != nil {
		status.OwnedResources = capsule.Status.OwnedResources
		status.UsedResources = capsule.Status.UsedResources
	}

	capsule.Status = status

	if err := r.pipeline.client.Status().Update(ctx, capsule); err != nil {
		return err
	}

	r.capsule.Status = status
	r.capsule.SetResourceVersion(capsule.GetResourceVersion())

	return nil
}

func normalizeObject(key objectKey, obj client.Object) client.Object {
	obj.SetManagedFields(nil)
	obj.GetObjectKind().SetGroupVersionKind(key.GroupVersionKind)
	return obj
}
