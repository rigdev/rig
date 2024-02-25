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
}

func newCapsuleRequest(p *Pipeline, capsule *v1alpha2.Capsule) *capsuleRequest {
	r := &capsuleRequest{
		pipeline: p,
		logger: p.logger.WithValues(
			"capsule", capsule.Name,
		),
		capsule:        capsule,
		currentObjects: map[objectKey]client.Object{},
	}

	if capsule.Status != nil {
		r.observedGeneration = capsule.Status.ObservedGeneration
	}

	r.logger.Info("created capsule request",
		"generation", capsule.Generation,
		"observed_generation", r.observedGeneration,
		"resource_version", capsule.ResourceVersion)

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

func (r *capsuleRequest) prepare() {
	r.usedResources = nil
	r.objects = map[objectKey]*Object{}
	for k, o := range r.currentObjects {
		r.objects[k] = &Object{
			Current: o.DeepCopyObject().(client.Object),
		}
	}
}

type change struct {
	state   resourceState
	applied bool
	err     error
}

func (r *capsuleRequest) commit(ctx context.Context) error {
	allKeys := maps.Keys(r.objects)

	// Prepare all the new objects with default labels / owner refs.
	for _, key := range allKeys {
		obj := r.objects[key]
		if obj.New == nil {
			continue
		}

		labels := obj.New.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[LabelOwnedByCapsule] = r.capsule.Name
		obj.New.SetLabels(labels)

		if err := controllerutil.SetControllerReference(r.capsule, obj.New, r.pipeline.scheme); err != nil {
			return err
		}
	}

	changes := map[objectKey]*change{}

	// Dry run to detect no-op vs create vs update.
	for _, key := range allKeys {
		obj := r.objects[key]

		if obj.Current == nil {
			materializedObj := obj.New.DeepCopyObject().(client.Object)
			if err := r.pipeline.client.Create(ctx, materializedObj, client.DryRunAll); kerrors.IsConflict(err) {
				return errors.FailedPreconditionErrorf("new object version available for '%v'", key)
			} else if kerrors.IsAlreadyExists(err) {
				o, err2 := r.pipeline.scheme.New(key.GroupVersionKind)
				if err2 != nil {
					return err
				}

				co := o.(client.Object)
				if err := r.pipeline.client.Get(ctx, key.ObjectKey, co); err != nil {
					return fmt.Errorf("could not get existing object: %w", err)
				}

				if IsOwnedBy(r.capsule, co) {
					r.logger.Info("object exists but not in status, retrying", "object", key)
					r.currentObjects[key] = co
					return errors.AbortedErrorf("object exists but not in capsule status")
				}

				r.logger.Info("create object skipped, not owned by controller", "object", key)
				changes[key] = &change{state: _resourceStateAlreadyExists}
				continue
			} else if err != nil {
				return fmt.Errorf("could not render create to %s: %w", key.GroupVersionKind, err)
			}

			r.logger.Info("create object", "object", key)
			changes[key] = &change{state: _resourceStateCreated}
			continue
		}

		if !IsOwnedBy(r.capsule, obj.Current) {
			r.logger.Info("update object skipped, not owned by controller", "object", key)
			changes[key] = &change{state: _resourceStateAlreadyExists}
			continue
		}

		if obj.New == nil {
			r.logger.Info("delete object", "object", key)
			changes[key] = &change{state: _resourceStateDeleted}
			continue
		}

		materializedObj := obj.New.DeepCopyObject().(client.Object)
		materializedObj.GetObjectKind().SetGroupVersionKind(obj.Current.GetObjectKind().GroupVersionKind())

		// Dry run to fully materialize the new spec.
		materializedObj.SetResourceVersion(obj.Current.GetResourceVersion())
		if err := r.pipeline.client.Update(ctx, materializedObj, client.DryRunAll); kerrors.IsConflict(err) {
			return errors.FailedPreconditionErrorf("new object version available for '%v'", key)
		} else if err != nil {
			return fmt.Errorf("could not render update to %s: %w", key.GroupVersionKind, err)
		}

		if ObjectsEquals(obj.Current, materializedObj) {
			r.logger.Info("update object skipped, not changed", "object", key)
			changes[key] = &change{state: _resourceStateUnchanged}
			continue
		}

		r.logger.Info("update object", "object", key)
		changes[key] = &change{state: _resourceStateUpdated}
	}

	// Skip update if no changes.
	if r.observedGeneration == r.capsule.Generation {
		r.logger.Info("already at generation", "generation", r.observedGeneration)
		hasChanges := false
		for _, change := range changes {
			switch change.state {
			case _resourceStateUpdated, _resourceStateCreated, _resourceStateDeleted:
				hasChanges = true
			}
		}
		if !hasChanges {
			r.logger.Info("no changes to apply", "generation", r.observedGeneration)
			return nil
		}
	}

	if err := r.updateStatusChanges(ctx, changes, r.observedGeneration); err != nil {
		return err
	}

	var errs []error
	for key, change := range changes {
		if err := r.applyChange(ctx, key, change.state); err != nil {
			change.err = err
			errs = append(errs, err)
		} else {
			change.applied = true
		}
	}

	if err := errors.Join(errs...); err != nil {
		return err
	}

	if err := r.updateStatusChanges(ctx, changes, r.capsule.Generation); err != nil {
		return err
	}

	return nil
}

func (r *capsuleRequest) applyChange(ctx context.Context, key objectKey, state resourceState) error {
	switch state {
	case _resourceStateUpdated:
		r.logger.Info("update object", "object", key)
		obj := r.objects[key]
		obj.New.SetResourceVersion(obj.Current.GetResourceVersion())
		if err := r.pipeline.client.Update(ctx, obj.New); err != nil {
			return fmt.Errorf("could not update %s: %w", key.GroupVersionKind, err)
		}

	case _resourceStateCreated:
		r.logger.Info("create object", "object", key)
		obj := r.objects[key]
		if err := r.pipeline.client.Create(ctx, obj.New); err != nil {
			return fmt.Errorf("could not create %s: %w", key.GroupVersionKind, err)
		}

	case _resourceStateDeleted:
		r.logger.Info("delete object", "object", key)
		obj := r.objects[key]
		if err := r.pipeline.client.Delete(ctx, obj.Current); err != nil {
			return fmt.Errorf("could not update %s: %w", key.GroupVersionKind, err)
		}
	}

	return nil
}

type resourceState string

const (
	_resourceStateDeleted       resourceState = "deleted"
	_resourceStateUpdated       resourceState = "updated"
	_resourceStateUnchanged     resourceState = "unchanged"
	_resourceStateCreated       resourceState = "created"
	_resourceStateFailed        resourceState = "failed"
	_resourceStateAlreadyExists resourceState = "alreadyExists"
	_resourceStateChangePending resourceState = "changePending"
)

func (r *capsuleRequest) updateStatusChanges(ctx context.Context, changes map[objectKey]*change, generation int64) error {
	capsule := r.capsule.DeepCopy()

	status := &v1alpha2.CapsuleStatus{
		ObservedGeneration: generation,
	}

	keys := maps.Keys(changes)
	slices.SortStableFunc(keys, func(k1, k2 objectKey) int { return strings.Compare(k1.String(), k2.String()) })
	for _, key := range keys {
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
		case _resourceStateCreated, _resourceStateUpdated, _resourceStateDeleted:
			if !change.applied {
				or.State = string(_resourceStateChangePending)
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
