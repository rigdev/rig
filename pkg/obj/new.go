package obj

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(gvk schema.GroupVersionKind, scheme *runtime.Scheme) client.Object {
	ro, err := scheme.New(gvk)
	if err != nil {
		ro = &unstructured.Unstructured{}
	}

	co := ro.(client.Object)

	co.GetObjectKind().SetGroupVersionKind(gvk)

	return co
}
