package pipeline

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/scheme"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//nolint:lll
type Request interface {
	// Scheme returns the serialization scheme used by the rig operator.
	// It contains all the types used by a Capsule.
	Scheme() *runtime.Scheme
	// Reader is a Kubernetes reader with access to the cluster the rig operator is running in.
	Reader() client.Reader
	// GetExisting populates 'obj' with a copy of the corresponding object owned by the capsule currently present in the cluster.
	// If the name of 'obj' isn't set, it defaults to the Capsule name.
	GetExisting(gvk schema.GroupVersionKind, name string) (client.Object, error)
	GetExistingInto(obj client.Object) error
	// GetNew populates 'obj' with a copy of the corresponding object owned by the capsule about to be applied.
	// If the name of 'obj' isn't set, it defaults to the Capsule name.
	GetNew(gvk schema.GroupVersionKind, name string) (client.Object, error)
	GetNewInto(obj client.Object) error
	// Set updates the object recorded to be applied.
	// If the name of 'obj' isn't set, it defaults to the Capsule name.
	Set(obj client.Object) error
	// Delete records the given object to be deleted.
	// The behavior is such that that calling req.Delete(obj) and then req.GetNew(obj)
	// returns a not-found error from GetNew.
	// If an object of the given type and name is present in the cluster, calling req.GetExisting(obj) succeeds
	// as calls to Delete (or Set) will only be applied to the cluster at the very end of the reconcilliation.
	// If the name of 'obj' isn't set, it defaults to the Capsule name.
	Delete(gvk schema.GroupVersionKind, name string) error
	// ListExisting returns a list with a copy of the objects of the corresponding type owned by the capsule and currently present in the cluster.
	// If you want a slice of typed objects, use the generic free-standing ListExisting function.
	ListExisting(gvk schema.GroupVersionKind) ([]client.Object, error)
	// ListNew returns a list with a copy of the objects of the corresponding type owned by the capsule and about to be applied.
	// If you want a slice of typed objects, use the generic free-standing ListNew function.
	ListNew(gvk schema.GroupVersionKind) ([]client.Object, error)
}

func ListExisting[T client.Object](r Request, obj T) ([]T, error) {
	gvks, _, err := r.Scheme().ObjectKinds(obj)
	if err != nil {
		return nil, err
	}

	objects, err := r.ListExisting(gvks[0])
	if err != nil {
		return nil, err
	}

	return ListConvert[T](objects)
}

func ListNew[T client.Object](r Request, obj T) ([]T, error) {
	gvks, _, err := r.Scheme().ObjectKinds(obj)
	if err != nil {
		return nil, err
	}

	objects, err := r.ListNew(gvks[0])
	if err != nil {
		return nil, err
	}
	return ListConvert[T](objects)
}

func ListConvert[T client.Object](objects []client.Object) ([]T, error) {
	var res []T
	for _, obj := range objects {
		o, ok := obj.(T)
		if !ok {
			return nil, fmt.Errorf("object had wrong type %T", obj)
		}
		res = append(res, o)
	}

	return res, nil
}

type RequestDeps struct {
	client client.Client
	reader client.Reader
	vm     scheme.VersionMapper
	config *configv1alpha1.OperatorConfig
	scheme *runtime.Scheme
	logger logr.Logger
}

// TODO Make generic over object?
type RequestState struct {
	requestObject      client.Object
	existingObjects    map[ObjectKey]client.Object
	newObjects         map[ObjectKey]*Object
	observedGeneration int64
	lastErrors         []string
	dryRun             bool
	force              bool
}

type RequestStrategies interface {
	// Status updating strategies
	UpdateStatusWithChanges(ctx context.Context, changes map[ObjectKey]*Change, generation int64) error
	UpdateStatusWithError(ctx context.Context, err error) error

	// Execution loop strategies
	LoadExistingObjects(ctx context.Context) error
	Prepare()
	OwnedLabel() string

	GetKey(gvk schema.GroupVersionKind, name string) (ObjectKey, error)
}

type RequestBase struct {
	RequestDeps
	RequestState
	Strategies RequestStrategies
}

func NewRequestBase(
	c client.Client,
	reader client.Reader,
	vm scheme.VersionMapper,
	config *configv1alpha1.OperatorConfig,
	scheme *runtime.Scheme,
	logger logr.Logger,
	strategies RequestStrategies,
	object client.Object,
) RequestBase {
	return RequestBase{
		RequestDeps: RequestDeps{
			client: c,
			reader: reader,
			vm:     vm,
			config: config,
			scheme: scheme,
			logger: logger,
		},
		RequestState: RequestState{
			existingObjects: map[ObjectKey]client.Object{},
			newObjects:      map[ObjectKey]*Object{},
			requestObject:   object,
		},
		Strategies: strategies,
	}
}

func (r *RequestBase) Scheme() *runtime.Scheme {
	return r.scheme
}

func (r *RequestBase) Reader() client.Reader {
	return r.reader
}

func (r *RequestBase) GetExisting(gvk schema.GroupVersionKind, name string) (client.Object, error) {
	key, err := r.Strategies.GetKey(gvk, name)
	if err != nil {
		return nil, err
	}

	o, ok := r.newObjects[key]
	if !ok {
		return nil, errors.NotFoundErrorf("object '%v' of type '%v' not found", key.Name, key.GroupVersionKind)
	}

	if o.Current == nil {
		return nil, errors.NotFoundErrorf("object '%v' of type '%v' has no existing version", key.Name, key.GroupVersionKind)
	}

	return o.Current.DeepCopyObject().(client.Object), nil
}

func (r *RequestBase) GetExistingInto(obj client.Object) error {
	gvk, err := getGVK(obj, r.scheme)
	if err != nil {
		return err
	}

	res, err := r.GetExisting(gvk, obj.GetName())
	if err != nil {
		return err
	}

	return r.scheme.Convert(res, obj, nil)
}

func (r *RequestBase) GetNew(gvk schema.GroupVersionKind, name string) (client.Object, error) {
	key, err := r.Strategies.GetKey(gvk, name)
	if err != nil {
		return nil, err
	}

	o, ok := r.newObjects[key]
	if !ok {
		return nil, errors.NotFoundErrorf("object '%v' of type '%v' not found", key.Name, key.GroupVersionKind)
	}

	if o.New == nil {
		return nil, errors.NotFoundErrorf("object '%v' of type '%v' has no new version", key.Name, key.GroupVersionKind)
	}

	return o.New.DeepCopyObject().(client.Object), nil
}

func (r *RequestBase) GetNewInto(obj client.Object) error {
	gvk, err := getGVK(obj, r.scheme)
	if err != nil {
		return err
	}

	res, err := r.GetNew(gvk, obj.GetName())
	if err != nil {
		return err
	}

	if err := r.scheme.Convert(res, obj, nil); err != nil {
		return err
	}

	return nil
}

func (r *RequestBase) Set(obj client.Object) error {
	gvk, err := getGVK(obj, r.scheme)
	if err != nil {
		return err
	}

	key, err := r.Strategies.GetKey(gvk, obj.GetName())
	if err != nil {
		return err
	}

	obj.SetName(key.Name)
	obj.SetNamespace(key.Namespace)

	o, ok := r.newObjects[key]
	if !ok {
		o = &Object{}
	}
	o.New = obj
	r.newObjects[key] = o
	return nil
}

func (r *RequestBase) Delete(gvk schema.GroupVersionKind, name string) error {
	key, err := r.Strategies.GetKey(gvk, name)
	if err != nil {
		return err
	}

	o, ok := r.newObjects[key]
	if ok {
		o.New = nil
	}

	return nil
}

func (r *RequestBase) ListExisting(gvk schema.GroupVersionKind) ([]client.Object, error) {
	gvk, err := r.getGVK(gvk)
	if err != nil {
		r.logger.Error(err, "invalid object list type")
		return nil, err
	}

	var res []client.Object
	for _, key := range sortedKeys(maps.Keys(r.existingObjects)) {
		if key.GroupVersionKind != gvk {
			continue
		}
		o := r.existingObjects[key].DeepCopyObject().(client.Object)
		res = append(res, o)
	}

	return res, nil
}

func (r *RequestBase) ListNew(gvk schema.GroupVersionKind) ([]client.Object, error) {
	gvk, err := r.getGVK(gvk)
	if err != nil {
		r.logger.Error(err, "invalid object list type")
		return nil, err
	}

	var res []client.Object
	for _, key := range sortedKeys(maps.Keys(r.newObjects)) {
		if key.GroupVersionKind != gvk {
			continue
		}
		no := r.newObjects[key]
		if no.New == nil {
			continue
		}
		o := no.New.DeepCopyObject().(client.Object)
		res = append(res, o)
	}

	return res, nil
}

func (r *RequestBase) Commit(ctx context.Context) (map[ObjectKey]*Change, error) {
	allKeys := sortedKeys(maps.Keys(r.newObjects))

	// Prepare all the new objects with default labels / owner refs.
	for _, key := range allKeys {
		cObj := r.newObjects[key]
		if cObj.New == nil {
			continue
		}

		cObj.New.GetObjectKind().SetGroupVersionKind(key.GroupVersionKind)
		normalizeObject(key, cObj.New)

		labels := cObj.New.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[r.Strategies.OwnedLabel()] = r.requestObject.GetName()
		cObj.New.SetLabels(labels)

		if err := controllerutil.SetControllerReference(r.requestObject, cObj.New, r.scheme); err != nil {
			return nil, fmt.Errorf("could not set controller ref: %q", err)
		}
	}

	changes := map[ObjectKey]*Change{}

	// Dry run to detect no-op vs create vs update.
	for _, key := range allKeys {
		cObj := r.newObjects[key]

		if cObj.Current == nil {
			materializedObj := cObj.New.DeepCopyObject().(client.Object)
			if err := errors.FromK8sClient(r.client.Create(ctx, materializedObj, client.DryRunAll)); errors.IsAborted(err) {
				return nil, errors.FailedPreconditionErrorf("new object version available for '%v'", key)
			} else if errors.IsAlreadyExists(err) || errors.IsInvalidArgument(err) {
				o, newErr := r.scheme.New(key.GroupVersionKind)
				if newErr != nil {
					return nil, err
				}

				co := o.(client.Object)
				if getErr := errors.FromK8sClient(r.client.Get(ctx, key.ObjectKey, co)); errors.IsNotFound(getErr) {
					r.logger.Info("configuration is invalid", "object", key, "error", err)
					return nil, err
				} else if getErr != nil {
					return nil, fmt.Errorf("could not get existing object: %w", getErr)
				}

				if r.force || IsOwnedBy(r.requestObject, co) {
					r.logger.Info("object exists but not in status, retrying", "object", key)
					r.existingObjects[key] = normalizeObject(key, co)
					return nil, errors.AbortedErrorf("object exists but not in capsule status")
				}

				r.logger.Info("create object skipped, not owned by controller", "object", key)
				changes[key] = &Change{state: ResourceStateAlreadyExists}
				continue
			} else if errors.IsFailedPrecondition(err) {
				return nil, errors.InvalidArgumentErrorf("invalid create config: %s", errors.MessageOf(err))
			} else if err != nil {
				return nil, fmt.Errorf("could not render create to %s: %w", key, err)
			}

			cObj.Materialized = normalizeObject(key, materializedObj)

			r.logger.Info("create object", "object", key)
			changes[key] = &Change{state: ResourceStateCreated}
			continue
		}

		if !(r.force || IsOwnedBy(r.requestObject, cObj.Current)) {
			r.logger.Info("update object skipped, not owned by controller", "object", key)
			changes[key] = &Change{state: ResourceStateAlreadyExists}
			continue
		}

		if cObj.New == nil {
			r.logger.Info("delete object", "object", key)
			changes[key] = &Change{state: ResourceStateDeleted}
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
		r.logger.Info("generating materialized version", "object", key, "resource_version", cObj.Current.GetResourceVersion())
		if err := r.client.Update(ctx, materializedObj, client.DryRunAll); kerrors.IsConflict(err) {
			// TODO(anders): This is duplicated, make a helper.
			o, newErr := r.scheme.New(key.GroupVersionKind)
			if newErr != nil {
				return nil, err
			}

			co := o.(client.Object)
			if getErr := r.client.Get(ctx, key.ObjectKey, co); kerrors.IsNotFound(getErr) {
				r.logger.Info("configuration is invalid", "object", key, "error", err)
				return nil, err
			} else if getErr != nil {
				return nil, fmt.Errorf("could not get existing object: %w", getErr)
			}

			if IsOwnedBy(r.requestObject, co) {
				r.logger.Info("current version changed, retrying", "object", key)
				r.existingObjects[key] = normalizeObject(key, co)
				return nil, errors.AbortedErrorf("object reload")
			}

			return nil, errors.FailedPreconditionErrorf("new object version available for '%v'", key)
		} else if err != nil {
			return nil, fmt.Errorf("could not render update to %s: %w", key, err)
		}

		equal, err := ObjectsEquals(cObj.Current, materializedObj, r.scheme)
		if err != nil {
			r.logger.Error(err, "equals failed",
				"current", cObj.Current.GetObjectKind().GroupVersionKind(),
				"new", materializedObj.GetObjectKind().GroupVersionKind())
			return nil, err
		}

		if equal {
			r.logger.Info("update object skipped, not changed", "object", key)
			changes[key] = &Change{state: ResourceStateUnchanged}
			continue
		}

		cObj.Materialized = normalizeObject(key, materializedObj)

		r.logger.Info("update object", "object", key)

		changes[key] = &Change{state: ResourceStateUpdated}
	}

	// Skip update if no changes.
	if r.observedGeneration == r.requestObject.GetGeneration() && len(r.lastErrors) == 0 {
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

	if err := r.Strategies.UpdateStatusWithChanges(ctx, changes, r.observedGeneration); err != nil {
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

	if err := r.Strategies.UpdateStatusWithChanges(ctx, changes, r.requestObject.GetGeneration()); err != nil {
		return nil, err
	}

	return changes, nil
}

func (r *RequestBase) applyChange(ctx context.Context, key ObjectKey, state ResourceState) error {
	switch state {
	case ResourceStateUpdated:
		r.logger.Info("update object", "object", key)
		obj := r.newObjects[key]

		// Edge case: Deployments, when they are scaled, will apply changes in the following order:
		//   1) Change the number of replicas (pods) in the current ReplicaSet
		//	 2) Create a new ReplicaSet is needed (Pod-template changes to the Deployment)
		//   2a) Delete old ReplicaSets.
		// Because we put an annotation on the Pod, we always have Pod-template change when applying
		// scale changes. And the result is that when scaling up, it will first match the new number
		// of replicas in the *old* ReplicaSet, before moving to the new one.
		// The "fix" is to apply the change in two rounds:
		// 	 1) First, we apply the template changes (use current number of replicas)
		//	 2) Then apply the scale changes.
		// That way the scale is only being applied to the new ReplicaSet.
		if key.GroupVersionKind == AppsDeploymentGVK && obj.Materialized != nil {
			currentObj := obj.Current.(*v1.Deployment)
			newObj := obj.New.(*v1.Deployment)
			materializedObj := obj.Materialized.(*v1.Deployment)
			if !equality.Semantic.DeepEqual(currentObj.Spec.Template, materializedObj.Spec.Template) {
				// We have changes to the template - don't apply potential replica changes yet.
				newObj.Spec.Replicas = currentObj.Spec.Replicas
			}
		}

		obj.New.SetResourceVersion(obj.Current.GetResourceVersion())
		if err := r.client.Update(ctx, obj.New); err != nil {
			return fmt.Errorf("could not update %s: %w", key.GroupVersionKind, err)
		}

	case ResourceStateCreated:
		r.logger.Info("create object", "object", key)
		obj := r.newObjects[key]
		if err := r.client.Create(ctx, obj.New); err != nil {
			return fmt.Errorf("could not create %s: %w", key.GroupVersionKind, err)
		}

	case ResourceStateDeleted:
		r.logger.Info("delete object", "object", key)
		obj := r.newObjects[key]
		if err := r.client.Delete(ctx, obj.Current); err != nil {
			return fmt.Errorf("could not update %s: %w", key.GroupVersionKind, err)
		}
	}

	return nil
}

func normalizeObject(key ObjectKey, obj client.Object) client.Object {
	obj.SetManagedFields(nil)
	obj.GetObjectKind().SetGroupVersionKind(key.GroupVersionKind)
	return obj
}

type Change struct {
	state   ResourceState
	applied bool
	err     error
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

func (r *RequestBase) PrepareRequest() *Result {
	result := &Result{}
	r.newObjects = map[ObjectKey]*Object{}
	for _, k := range sortedKeys(maps.Keys(r.existingObjects)) {
		o := r.existingObjects[k]
		r.newObjects[k] = &Object{
			Current: o.DeepCopyObject().(client.Object),
		}
		oo := o.DeepCopyObject()
		oo.GetObjectKind().SetGroupVersionKind(k.GroupVersionKind)
		result.InputObjects = append(result.InputObjects, oo.(client.Object))
	}

	r.Strategies.Prepare()

	return result
}

func (r *RequestBase) getGVK(gvk schema.GroupVersionKind) (schema.GroupVersionKind, error) {
	gvk2, err := r.vm.FromGroupKind(gvk.GroupKind())
	if err != nil {
		if gvk.Version == "" {
			return schema.GroupVersionKind{}, err
		}
		gvk2 = gvk
	}
	return gvk2, nil
}

func getGVK(obj runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		if obj.GetObjectKind().GroupVersionKind().Version == "" {
			return schema.GroupVersionKind{}, err
		}
		return obj.GetObjectKind().GroupVersionKind(), nil
	}
	return gvks[0], nil
}
