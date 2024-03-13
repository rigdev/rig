package roclient

import (
	"context"
	gojson "encoding/json"

	"github.com/rigdev/rig/pkg/scheme"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	opts ...client.GetOption) error {
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

	s := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme,
		scheme,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
		},
	)

	for _, reader := range r.readers {
		newList := list.DeepCopyObject().(client.ObjectList)
		if err := reader.List(ctx, newList, opts...); err != nil {
			return err
		}

		listBytes, err := gojson.Marshal(newList)
		if err != nil {
			return err
		}

		list := struct {
			Items []gojson.RawMessage `json:"items"`
		}{}
		if err := gojson.Unmarshal(listBytes, &list); err != nil {
			return err
		}

		for _, item := range list.Items {
			ro, gvk, err := s.Decode(item, nil, nil)
			if err != nil {
				return err
			}

			ro.GetObjectKind().SetGroupVersionKind(*gvk)
			co := ro.(client.Object)

			found := false
			for _, o := range objs {
				if o.GetName() == co.GetName() &&
					o.GetObjectKind().GroupVersionKind().GroupKind() == co.GetObjectKind().GroupVersionKind().GroupKind() &&
					o.GetNamespace() == co.GetNamespace() {
					found = true
					break
				}
			}
			if !found {
				objs = append(objs, object{
					Object: co,
					raw:    item,
				})
			}
		}
	}

	bs, err := gojson.Marshal(map[string]interface{}{
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
