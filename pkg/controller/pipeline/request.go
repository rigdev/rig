package pipeline

import (
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
