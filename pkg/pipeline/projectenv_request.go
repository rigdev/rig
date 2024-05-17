package pipeline

import (
	"context"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/scheme"
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

type ProjectEnvironmentRequest interface {
	Request
	// ProjectEnvironment returns a deepcopy of the capsule object being reconciled.
	ProjectEnvironment() *v1alpha2.ProjectEnvironment
}

type projectEnvRequest struct {
	RequestBase
	projectEnv *v1alpha2.ProjectEnvironment
}

func NewProjectEnvironmentRequest(
	c client.Client,
	reader client.Reader,
	vm scheme.VersionMapper,
	config *configv1alpha1.OperatorConfig,
	scheme *runtime.Scheme,
	logger logr.Logger,
	projectEnv *v1alpha2.ProjectEnvironment,
) ExecutableRequest[ProjectEnvironmentRequest] {
	p := &projectEnvRequest{
		RequestBase: NewRequestBase(c, reader, vm, config, scheme, logger, nil, projectEnv),
		projectEnv:  projectEnv,
	}
	// TODO Fix this hack
	p.Strategies = p

	if projectEnv.Status != nil {
		p.observedGeneration = projectEnv.Status.ObservedGeneration
	}

	return p
}

func (p *projectEnvRequest) ProjectEnvironment() *v1alpha2.ProjectEnvironment {
	return p.projectEnv.DeepCopy()
}

func (p *projectEnvRequest) GetKey(gk schema.GroupKind, name string) (ObjectKey, error) {
	gvk, err := p.getGVK(gk)
	if err != nil {
		return ObjectKey{}, err
	}

	return ObjectKey{
		ObjectKey: types.NamespacedName{
			Name: name,
		},
		GroupVersionKind: gvk,
	}, nil
}

func (p *projectEnvRequest) LoadExistingObjects(ctx context.Context) error {
	s := p.projectEnv.Status
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

		gvk, err := p.vm.FromGroupKind(gk)
		if err != nil {
			return err
		}

		co := obj.New(gvk, p.scheme)

		_, isNamespace := co.(*corev1.Namespace)

		co.SetName(o.Ref.Name)
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

func (p *projectEnvRequest) UpdateStatusWithChanges(
	ctx context.Context,
	changes map[ObjectKey]*Change,
	generation int64,
) error {
	projectCopy := p.projectEnv.DeepCopy()

	status := &v1alpha2.ProjectEnvironmentStatus{
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
			if change.state == ResourceStateCreated {
				if key.Kind == "Namespace" && key.Name == p.requestObject.GetName() {
					status.CreatedNamespace = true
				}
			}
		}
		if change.err != nil {
			or.Message = change.err.Error()
		}
		status.OwnedResources = append(status.OwnedResources, or)
	}
	status.CreatedNamespace = status.CreatedNamespace || projectCopy.Status.CreatedNamespace
	projectCopy.Status = status
	if err := p.client.Status().Update(ctx, projectCopy); err != nil {
		return err
	}

	p.observedGeneration = generation
	p.projectEnv.Status = status
	p.projectEnv.SetResourceVersion(p.projectEnv.GetResourceVersion())

	return nil
}

func (p *projectEnvRequest) UpdateStatusWithError(ctx context.Context, err error) error {
	projectCopy := p.projectEnv.DeepCopy()

	status := &v1alpha2.ProjectEnvironmentStatus{
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

	p.projectEnv.Status = status
	p.projectEnv.SetResourceVersion(projectCopy.GetResourceVersion())

	return nil
}

func (*projectEnvRequest) Prepare() {}

func (*projectEnvRequest) OwnedLabel() string {
	return LabelOwnedByProject
}

func (p *projectEnvRequest) GetBase() *RequestBase {
	return &p.RequestBase
}

func (p *projectEnvRequest) GetRequest() ProjectEnvironmentRequest {
	return p
}
