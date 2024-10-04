package roclient

import (
	"context"
	gojson "encoding/json"
	"strings"

	"github.com/rigdev/rig/pkg/scheme"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type LayeredReader struct {
	readers []client.Reader
}

func NewLayeredReader(rs ...client.Reader) client.Reader {
	return &LayeredReader{readers: rs}
}

func (r *LayeredReader) Get(ctx context.Context,
	key client.ObjectKey,
	obj client.Object,
	opts ...client.GetOption,
) error {
	for _, reader := range r.readers {
		if err := reader.Get(ctx, key, obj, opts...); err == nil {
			return nil
		}
	}

	return kerrors.NewNotFound(schema.GroupResource{Group: obj.GetObjectKind().GroupVersionKind().Group}, key.Name)
}

func (r *LayeredReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	objs := []client.Object{}
	scheme := scheme.New()

	listGvk, err := apiutil.GVKForObject(list, scheme)
	if err != nil {
		return err
	}

	objGvk := listGvk
	objGvk.Kind = strings.TrimSuffix(objGvk.Kind, "List")

	for _, reader := range r.readers {
		newList := &unstructured.UnstructuredList{}
		newList.GetObjectKind().SetGroupVersionKind(listGvk)
		if err := reader.List(ctx, newList, opts...); err != nil {
			return err
		}

		for _, item := range newList.Items {
			item := &item
			if item.GetObjectKind().GroupVersionKind().Empty() {
				item.GetObjectKind().SetGroupVersionKind(objGvk)
			}

			found := false
			for _, o := range objs {
				if o.GetName() == item.GetName() &&
					o.GetObjectKind().GroupVersionKind().GroupKind() == item.GetObjectKind().GroupVersionKind().GroupKind() &&
					o.GetNamespace() == item.GetNamespace() {
					found = true
					break
				}
			}
			if !found {
				objs = append(objs, item)
			}
		}
	}

	bs, err := gojson.Marshal(map[string]any{
		"items": objs,
	})
	if err != nil {
		return err
	}

	if err := gojson.Unmarshal(bs, list); err != nil {
		return err
	}

	return nil
}
