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

const (
	LabelOwnedByCapsule = "rig.dev/owned-by-capsule"
)

type Object struct {
	Current client.Object
	New     client.Object
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
	client             client.Client
	config             *configv1alpha1.OperatorConfig
	scheme             *runtime.Scheme
	logger             logr.Logger
	capsule            *v1alpha2.Capsule
	currentObjects     map[objectKey]client.Object
	objects            map[objectKey]*Object
	steps              []Step
	observedGeneration int64
	usedResources      []v1alpha2.UsedResource
}

func New(
	cc client.Client,
	config *configv1alpha1.OperatorConfig,
	capsule *v1alpha2.Capsule,
	scheme *runtime.Scheme,
	logger logr.Logger,
) *Pipeline {
	p := &Pipeline{
		client: cc,
		config: config,
		scheme: scheme,
		logger: logger.WithValues(
			"capsule", capsule.Name,
		),
		capsule:        capsule,
		currentObjects: map[objectKey]client.Object{},
	}
	if capsule.Status != nil {
		p.observedGeneration = capsule.Status.ObservedGeneration
	}
	p.logger.Info("created pipeline",
		"generation", capsule.Generation,
		"observed_generation", p.observedGeneration,
		"resource_version", capsule.ResourceVersion)
	return p
}

func (p *Pipeline) Config() *configv1alpha1.OperatorConfig {
	return p.config.DeepCopy()
}

func (p *Pipeline) Scheme() *runtime.Scheme {
	return p.scheme
}

func (p *Pipeline) Capsule() *v1alpha2.Capsule {
	return p.capsule.DeepCopy()
}

func (p *Pipeline) Client() client.Client {
	return p.client
}

func (p *Pipeline) GetCurrent(obj client.Object) error {
	key, err := p.getKey(obj)
	if err != nil {
		return err
	}

	o, ok := p.objects[key]
	if !ok {
		return errors.NotFoundErrorf("object '%v' of type '%v' not found", key.Name, key.GroupVersionKind)
	}

	if o.Current == nil {
		return errors.NotFoundErrorf("object '%v' of type '%v' has no existing version", key.Name, key.GroupVersionKind)
	}

	return p.scheme.Converter().Convert(o.Current, obj, nil)
}

func (p *Pipeline) GetNew(obj client.Object) error {
	key, err := p.getKey(obj)
	if err != nil {
		return err
	}

	o, ok := p.objects[key]
	if !ok {
		return errors.NotFoundErrorf("object '%v' of type '%v' not found", key.Name, key.GroupVersionKind)
	}

	if o.New == nil {
		return errors.NotFoundErrorf("object '%v' of type '%v' has no new version", key.Name, key.GroupVersionKind)
	}

	return p.scheme.Converter().Convert(o.New, obj, nil)
}

func (p *Pipeline) getGVK(obj client.Object) (schema.GroupVersionKind, error) {
	gvks, _, err := p.scheme.ObjectKinds(obj)
	if err != nil {
		p.logger.Error(err, "invalid object type")
		return schema.GroupVersionKind{}, err
	}

	return gvks[0], nil
}

func (p *Pipeline) getKey(obj client.Object) (objectKey, error) {
	if obj.GetName() == "" {
		obj.SetName(p.capsule.Name)
	}
	obj.SetNamespace(p.capsule.Namespace)

	gvk, err := p.getGVK(obj)
	if err != nil {
		return objectKey{}, err
	}

	obj.SetNamespace(p.capsule.Namespace)
	return p.namedObjectKey(obj.GetName(), gvk), nil
}

func (p *Pipeline) Set(obj client.Object) error {
	key, err := p.getKey(obj)
	if err != nil {
		return err
	}

	o, ok := p.objects[key]
	if !ok {
		o = &Object{}
	}
	o.New = obj
	p.objects[key] = o
	return nil
}

func (p *Pipeline) Delete(obj client.Object) error {
	key, err := p.getKey(obj)
	if err != nil {
		return err
	}

	o, ok := p.objects[key]
	if ok {
		o.New = nil
	}

	return nil
}

func (p *Pipeline) namedObjectKey(name string, gvk schema.GroupVersionKind) objectKey {
	return objectKey{
		ObjectKey: types.NamespacedName{
			Name:      name,
			Namespace: p.capsule.Namespace,
		},
		GroupVersionKind: gvk,
	}
}

func (p *Pipeline) MarkUsedResource(res v1alpha2.UsedResource) {
	p.usedResources = append(p.usedResources, res)
}

func (p *Pipeline) AddStep(step Step) {
	p.steps = append(p.steps, step)
}

func (p *Pipeline) Run(ctx context.Context) error {
	if err := p.runSteps(ctx); errors.IsFailedPrecondition(err) {
		return err
	} else if err != nil {
		if err := p.updateStatusError(ctx, err); err != nil {
			return err
		}

		return err
	}

	return nil
}

func (p *Pipeline) runSteps(ctx context.Context) error {
	// Read all status objects.
	if s := p.capsule.Status; s != nil {
		for _, r := range s.OwnedResources {
			if r.Ref == nil {
				continue
			}
			gk := schema.GroupKind{
				Kind: r.Ref.Kind,
			}
			if r.Ref.APIGroup != nil {
				gk.Group = *r.Ref.APIGroup
			}

			gvk, err := LookupGVK(gk)
			if err != nil {
				return err
			}

			o, err := p.scheme.New(gvk)
			if err != nil {
				return err
			}

			co, ok := o.(client.Object)
			if !ok {
				continue
			}

			co.SetName(r.Ref.Name)
			co.SetNamespace(p.capsule.Namespace)
			if err := p.client.Get(ctx, client.ObjectKeyFromObject(co), co); kerrors.IsNotFound(err) {
				// Okay it doesn't exist, ignore the resource.
				continue
			} else if err != nil {
				return err
			}

			p.currentObjects[p.namedObjectKey(r.Ref.Name, gvk)] = co
		}
	}

	for {
		p.usedResources = nil
		p.objects = map[objectKey]*Object{}
		for k, o := range p.currentObjects {
			p.objects[k] = &Object{
				Current: o.DeepCopyObject().(client.Object),
			}
		}

		p.logger.Info("run steps", "current_objects", maps.Keys(p.currentObjects))

		for _, s := range p.steps {
			if err := s.Apply(ctx, p); err != nil {
				return err
			}
		}

		if err := p.commit(ctx); errors.IsAborted(err) {
			p.logger.Error(err, "retry running steps")
			continue
		} else if err != nil {
			p.logger.Error(err, "error committing changes")
			return err
		}

		return nil
	}
}

type change struct {
	state   resourceState
	applied bool
	err     error
}

func (p *Pipeline) commit(ctx context.Context) error {
	allKeys := maps.Keys(p.objects)

	// Prepare all the new objects with default labels / owner refs.
	for _, key := range allKeys {
		obj := p.objects[key]
		if obj.New == nil {
			continue
		}

		labels := obj.New.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[LabelOwnedByCapsule] = p.capsule.Name
		obj.New.SetLabels(labels)

		if err := controllerutil.SetControllerReference(p.capsule, obj.New, p.scheme); err != nil {
			return err
		}
	}

	changes := map[objectKey]*change{}

	// Dry run to detect no-op vs create vs update.
	for _, key := range allKeys {
		obj := p.objects[key]

		if obj.Current == nil {
			materializedObj := obj.New.DeepCopyObject().(client.Object)
			if err := p.client.Create(ctx, materializedObj, client.DryRunAll); kerrors.IsConflict(err) {
				return errors.FailedPreconditionErrorf("new object version available for '%v'", key)
			} else if kerrors.IsAlreadyExists(err) {
				o, err2 := p.scheme.New(key.GroupVersionKind)
				if err2 != nil {
					return err
				}

				co := o.(client.Object)
				if err := p.client.Get(ctx, key.ObjectKey, co); err != nil {
					return fmt.Errorf("could not get existing object: %w", err)
				}

				if IsOwnedBy(p.capsule, co) {
					p.logger.Info("object exists but not in status, retrying", "object", key)
					p.currentObjects[key] = co
					return errors.AbortedErrorf("object exists but not in capsule status")
				}

				p.logger.Info("create object skipped, not owned by controller", "object", key)
				changes[key] = &change{state: _resourceStateAlreadyExists}
				continue
			} else if err != nil {
				return fmt.Errorf("could not render create to %s: %w", key.GroupVersionKind, err)
			}

			p.logger.Info("create object", "object", key)
			changes[key] = &change{state: _resourceStateCreated}
			continue
		}

		if !IsOwnedBy(p.capsule, obj.Current) {
			p.logger.Info("update object skipped, not owned by controller", "object", key)
			changes[key] = &change{state: _resourceStateAlreadyExists}
			continue
		}

		if obj.New == nil {
			p.logger.Info("delete object", "object", key)
			changes[key] = &change{state: _resourceStateDeleted}
			continue
		}

		materializedObj := obj.New.DeepCopyObject().(client.Object)
		materializedObj.GetObjectKind().SetGroupVersionKind(obj.Current.GetObjectKind().GroupVersionKind())

		// Dry run to fully materialize the new spec.
		materializedObj.SetResourceVersion(obj.Current.GetResourceVersion())
		if err := p.client.Update(ctx, materializedObj, client.DryRunAll); kerrors.IsConflict(err) {
			return errors.FailedPreconditionErrorf("new object version available for '%v'", key)
		} else if err != nil {
			return fmt.Errorf("could not render update to %s: %w", key.GroupVersionKind, err)
		}

		if ObjectsEquals(obj.Current, materializedObj) {
			p.logger.Info("update object skipped, not changed", "object", key)
			changes[key] = &change{state: _resourceStateUnchanged}
			continue
		}

		p.logger.Info("update object", "object", key)
		changes[key] = &change{state: _resourceStateUpdated}
	}

	// Skip update if no changes.
	if p.observedGeneration == p.capsule.Generation {
		p.logger.Info("already at generation", "generation", p.observedGeneration)
		hasChanges := false
		for _, change := range changes {
			switch change.state {
			case _resourceStateUpdated, _resourceStateCreated, _resourceStateDeleted:
				hasChanges = true
			}
		}
		if !hasChanges {
			p.logger.Info("no changes to apply", "generation", p.observedGeneration)
			return nil
		}
	}

	if err := p.updateStatusChanges(ctx, changes, p.observedGeneration); err != nil {
		return err
	}

	var errs []error
	for key, change := range changes {
		if err := p.applyChange(ctx, key, change.state); err != nil {
			change.err = err
			errs = append(errs, err)
		} else {
			change.applied = true
		}
	}

	if err := errors.Join(errs...); err != nil {
		return err
	}

	if err := p.updateStatusChanges(ctx, changes, p.capsule.Generation); err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) applyChange(ctx context.Context, key objectKey, state resourceState) error {
	switch state {
	case _resourceStateUpdated:
		p.logger.Info("update object", "object", key)
		obj := p.objects[key]
		obj.New.SetResourceVersion(obj.Current.GetResourceVersion())
		if err := p.client.Update(ctx, obj.New); err != nil {
			return fmt.Errorf("could not update %s: %w", key.GroupVersionKind, err)
		}

	case _resourceStateCreated:
		p.logger.Info("create object", "object", key)
		obj := p.objects[key]
		if err := p.client.Create(ctx, obj.New); err != nil {
			return fmt.Errorf("could not create %s: %w", key.GroupVersionKind, err)
		}

	case _resourceStateDeleted:
		p.logger.Info("delete object", "object", key)
		obj := p.objects[key]
		if err := p.client.Delete(ctx, obj.Current); err != nil {
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

func (p *Pipeline) updateStatusChanges(ctx context.Context, changes map[objectKey]*change, generation int64) error {
	capsule := p.capsule.DeepCopy()

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

	status.UsedResources = p.usedResources

	capsule.Status = status

	if err := p.client.Status().Update(ctx, capsule); err != nil {
		return err
	}

	p.observedGeneration = generation
	p.capsule.Status = status
	p.capsule.SetResourceVersion(capsule.GetResourceVersion())

	return nil
}

func (p *Pipeline) updateStatusError(ctx context.Context, err error) error {
	capsule := p.capsule.DeepCopy()

	status := &v1alpha2.CapsuleStatus{
		ObservedGeneration: p.observedGeneration,
		Errors:             []string{err.Error()},
	}

	if capsule.Status != nil {
		status.OwnedResources = capsule.Status.OwnedResources
		status.UsedResources = capsule.Status.UsedResources
	}

	capsule.Status = status

	if err := p.client.Status().Update(ctx, capsule); err != nil {
		return err
	}

	p.capsule.Status = status
	p.capsule.SetResourceVersion(capsule.GetResourceVersion())

	return nil
}
