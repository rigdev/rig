package controller

import (
	"context"
	"fmt"
	"reflect"
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

func GetNew[T client.Object](req Request, key ObjectKey) T {
	var t T
	o := req.GetNew(key)
	if c, ok := o.(T); ok {
		t = c
	}
	return t
}

func GetCurrent[T client.Object](req Request, key ObjectKey) T {
	var t T
	o := req.GetCurrent(key)
	if c, ok := o.(T); ok {
		t = c
	}
	return t
}

func Get[T interface {
	client.Object
	comparable
}](req Request, key ObjectKey,
) T {
	var t T
	if o := GetNew[T](req, key); o != t {
		return o
	}
	if o := GetCurrent[T](req, key); o != t {
		return o
	}
	return t
}

type Object struct {
	Current client.Object
	New     client.Object
}

type ObjectKey struct {
	client.ObjectKey
	schema.GroupVersionKind
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

type Request interface {
	Config() *configv1alpha1.OperatorConfig
	Scheme() *runtime.Scheme
	Client() client.Client
	Capsule() *v1alpha2.Capsule
	GetCurrent(key ObjectKey) client.Object
	GetNew(key ObjectKey) client.Object
	Set(key ObjectKey, obj client.Object)
	NamedObjectKey(name string, gvk schema.GroupVersionKind) ObjectKey
	ObjectKey(gvk schema.GroupVersionKind) ObjectKey
	MarkUsedResource(res v1alpha2.UsedResource)
}

type Step interface {
	Apply(ctx context.Context, req Request) error
}

type Pipeline struct {
	client         client.Client
	config         *configv1alpha1.OperatorConfig
	scheme         *runtime.Scheme
	logger         logr.Logger
	capsule        *v1alpha2.Capsule
	currentObjects map[ObjectKey]client.Object
	objects        map[ObjectKey]*Object
	steps          []Step
	generation     int64
	usedResources  []v1alpha2.UsedResource
}

func NewPipeline(
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
		currentObjects: map[ObjectKey]client.Object{},
	}
	if capsule.Status != nil {
		p.generation = capsule.Status.ObservedGeneration
	}
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

func (p *Pipeline) GetCurrent(key ObjectKey) client.Object {
	o, ok := p.objects[key]
	if !ok {
		return nil
	}

	if o.Current == nil {
		return nil
	}

	return o.Current.DeepCopyObject().(client.Object)
}

func (p *Pipeline) GetNew(key ObjectKey) client.Object {
	o, ok := p.objects[key]
	if !ok {
		return nil
	}

	if o.New == nil {
		return nil
	}

	return o.New.DeepCopyObject().(client.Object)
}

func (p *Pipeline) Set(key ObjectKey, obj client.Object) {
	o, ok := p.objects[key]
	if !ok {
		o = &Object{}
	}
	o.New = obj
	p.objects[key] = o
}

func (p *Pipeline) NamedObjectKey(name string, gvk schema.GroupVersionKind) ObjectKey {
	return ObjectKey{
		ObjectKey: types.NamespacedName{
			Name:      name,
			Namespace: p.capsule.Namespace,
		},
		GroupVersionKind: gvk,
	}
}

func (p *Pipeline) ObjectKey(gvk schema.GroupVersionKind) ObjectKey {
	return p.NamedObjectKey(p.capsule.Name, gvk)
}

func (p *Pipeline) MarkUsedResource(res v1alpha2.UsedResource) {
	p.usedResources = append(p.usedResources, res)
}

func (p *Pipeline) AddStep(step Step) {
	p.steps = append(p.steps, step)
}

func (p *Pipeline) Run(ctx context.Context) error {
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

			gvk, err := lookupGVK(gk)
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

			p.currentObjects[p.NamedObjectKey(r.Ref.Name, gvk)] = co
		}
	}

	for {
		p.usedResources = nil
		p.objects = map[ObjectKey]*Object{}
		for k, o := range p.currentObjects {
			p.objects[k] = &Object{
				Current: o.DeepCopyObject().(client.Object),
			}
		}

		p.logger.Info("run steps", "current_objects", maps.Keys(p.currentObjects))

		for _, s := range p.steps {
			if err := s.Apply(ctx, p); err != nil {
				if err := p.updateStatus(ctx, nil, err); err != nil {
					return err
				}
				return err
			}
		}

		if err := p.commit(ctx); errors.IsAborted(err) {
			p.logger.Error(err, "retry running steps")
			continue
		} else if err != nil {
			p.logger.Error(err, "error committing changes")
			if err := p.updateStatus(ctx, nil, err); err != nil {
				return err
			}

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

	changes := map[ObjectKey]*change{}

	// Dry run to detect no-op vs create vs update.
	for _, key := range allKeys {
		obj := p.objects[key]

		if obj.Current == nil {
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
		if err := p.client.Update(ctx, materializedObj, client.DryRunAll); err != nil {
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
	if p.generation == p.capsule.Generation {
		p.logger.Info("already at generation", "generation", p.generation)
		hasChanges := false
		for _, change := range changes {
			switch change.state {
			case _resourceStateUpdated, _resourceStateCreated, _resourceStateDeleted:
				hasChanges = true
			}
		}
		if !hasChanges {
			p.logger.Info("no changes to apply", "generation", p.generation)
			return nil
		}
	}

	if err := p.updateStatus(ctx, changes, nil); err != nil {
		return err
	}

	var errs []error
	for key, change := range changes {
		if err := p.applyChange(ctx, key, change.state); errors.IsAborted(err) {
			return err
		} else if err != nil {
			change.err = err
			errs = append(errs, err)
		} else {
			change.applied = true
		}
	}

	err := errors.Join(errs...)
	if err == nil {
		p.generation = p.capsule.Generation
	}
	// If there is an error, it's set here on the status (or failing trying).
	if err := p.updateStatus(ctx, changes, err); err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) applyChange(ctx context.Context, key ObjectKey, state resourceState) error {
	switch state {
	case _resourceStateUpdated:
		p.logger.Info("update object", "object", key)
		obj := p.objects[key]
		if err := p.client.Update(ctx, obj.New); err != nil {
			return fmt.Errorf("could not update %s: %w", key.GroupVersionKind, err)
		}

	case _resourceStateCreated:
		p.logger.Info("create object", "object", key)
		obj := p.objects[key]
		if err := p.client.Create(ctx, obj.New); err != nil {
			o, err2 := p.scheme.New(key.GroupVersionKind)
			if err2 != nil {
				return err
			}

			co, ok := o.(client.Object)
			if !ok {
				return fmt.Errorf("invalid object conversion: %v", reflect.TypeOf(o))
			}

			if err2 := p.client.Get(ctx, key.ObjectKey, co); err2 == nil {
				if IsOwnedBy(p.capsule, co) {
					p.logger.Info("object exists but not in status, retrying", "object", key)
					p.currentObjects[key] = co
					return errors.AbortedErrorf("object exists but not in capsule status")
				}
			}

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

func (p *Pipeline) updateStatus(ctx context.Context, changes map[ObjectKey]*change, err error) error {
	capsule := p.capsule.DeepCopy()

	status := &v1alpha2.CapsuleStatus{
		ObservedGeneration: p.generation,
	}

	if changes != nil {
		keys := maps.Keys(changes)
		slices.SortStableFunc(keys, func(k1, k2 ObjectKey) int { return strings.Compare(k1.String(), k2.String()) })
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
	} else {
		status.OwnedResources = capsule.Status.OwnedResources
	}

	if err != nil {
		status.Errors = []string{err.Error()}
	}

	status.UsedResources = p.usedResources

	capsule.Status = status

	if err := p.client.Status().Update(ctx, capsule); err != nil {
		return err
	}

	p.capsule.Status = status
	p.capsule.SetResourceVersion(capsule.GetResourceVersion())

	return nil
}
