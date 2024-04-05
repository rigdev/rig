package pipeline

import (
	"context"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LabelOwnedByProject = "rig.dev/owned-by-project"
)

type ProjectRequest interface {
	Request
	// Project returns a deepcopy of the capsule object being reconciled.
	Project() *v1alpha2.Project
}

type projectRequest struct {
	RequestBase
	project *v1alpha2.Project
}

func NewProjectRequest(
	c client.Client,
	reader client.Reader,
	config *configv1alpha1.OperatorConfig,
	scheme *runtime.Scheme,
	logger logr.Logger,
	project *v1alpha2.Project,
) ExecutableRequest[ProjectRequest] {
	p := &projectRequest{
		RequestBase: NewRequestBase(c, reader, config, scheme, logger, nil, project),
		project:     project,
	}
	// TODO Fix this hack
	p.Strategies = p

	if project.Status != nil {
		p.observedGeneration = project.Status.ObservedGeneration
	}

	return p
}

func (p *projectRequest) Project() *v1alpha2.Project {
	return p.project.DeepCopy()
}

func (p *projectRequest) GetKey(obj client.Object) (ObjectKey, error) {
	gvk, err := getGVK(obj, p.scheme)
	if err != nil {
		p.logger.Error(err, "invalid object type")
		return ObjectKey{}, err
	}

	ns := obj.GetNamespace()
	if _, ok := obj.(*corev1.Namespace); ok {
		ns = obj.GetName()
	} else if obj.GetName() == "" {
		obj.SetName(obj.GetNamespace())
	}

	return ObjectKey{
		ObjectKey: types.NamespacedName{
			Namespace: obj.GetName(),
			Name:      ns,
		},
		GroupVersionKind: gvk,
	}, nil
}

func (p *projectRequest) LoadExistingObjects(ctx context.Context) error {
	s := p.project.Status
	if s == nil {
		return nil
	}

	for _, o := range s.OwnedResources {
		if o.Ref == nil {
			continue
		}
		gk := schema.GroupKind{Kind: o.Ref.Kind}
		if o.Ref.APIGroup != nil {
			gk.Group = *o.Ref.APIGroup
		}

		gvk, err := LookupGVK(gk)
		if err != nil {
			return err
		}

		ro, err := p.scheme.New(gvk)
		if err != nil {
			return err
		}

		co, ok := ro.(client.Object)
		if !ok {
			continue
		}

		_, isNamespace := co.(*corev1.Namespace)

		co.SetName(o.Ref.Name)
		co.GetObjectKind().SetGroupVersionKind(gvk)
		if !isNamespace {
			co.SetNamespace(*o.Ref.Namespace)
		}
		if err := p.reader.Get(ctx, client.ObjectKeyFromObject(co), co); kerrors.IsNotFound(err) {
			// Okay it doesn't exist, ignore the resource.
			continue
		} else if err != nil {
			return err
		}

		ns := co.GetNamespace()
		if isNamespace {
			ns = co.GetName()
		}
		p.existingObjects[ObjectKey{
			ObjectKey: types.NamespacedName{
				Namespace: ns,
				Name:      co.GetName(),
			},
			GroupVersionKind: gvk,
		}] = co
	}

	return nil
}

func (p *projectRequest) UpdateStatusWithChanges(
	ctx context.Context,
	changes map[ObjectKey]*Change,
	generation int64,
) error {
	projectCopy := p.project.DeepCopy()

	status := &v1alpha2.ProjectStatus{
		ObservedGeneration: generation,
	}

	for _, key := range sortedKeys(maps.Keys(changes)) {
		key := key
		change := changes[key]
		or := v1alpha2.OwnedGlobalResource{
			Ref: &corev1.TypedObjectReference{
				APIGroup:  &key.Group,
				Kind:      key.Kind,
				Name:      key.Name,
				Namespace: &key.Namespace,
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
	projectCopy.Status = status
	if err := p.client.Status().Update(ctx, projectCopy); err != nil {
		return err
	}

	p.observedGeneration = generation
	p.project.Status = status
	p.project.SetResourceVersion(p.project.GetResourceVersion())

	return nil
}

func (p *projectRequest) UpdateStatusWithError(ctx context.Context, err error) error {
	projectCopy := p.project.DeepCopy()

	status := &v1alpha2.ProjectStatus{
		ObservedGeneration: p.observedGeneration,
		Errors:             []string{err.Error()},
	}

	if projectCopy.Status != nil {
		status.OwnedResources = projectCopy.Status.OwnedResources
	}
	projectCopy.Status = status

	if err := p.client.Status().Update(ctx, projectCopy); err != nil {
		return err
	}

	p.project.Status = status
	p.project.SetResourceVersion(projectCopy.GetResourceVersion())

	return nil
}

func (*projectRequest) Prepare() {}

func (*projectRequest) OwnedLabel() string {
	return LabelOwnedByProject
}

func (p *projectRequest) GetBase() *RequestBase {
	return &p.RequestBase
}

func (p *projectRequest) GetRequest() ProjectRequest {
	return p
}
