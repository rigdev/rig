package pipeline

import (
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
