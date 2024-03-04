package roclient

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"io"
	"os"
	"strings"

	"github.com/rigdev/rig/pkg/errors"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"
)

type Reader interface {
	client.Reader

	AddObject(co client.Object) error
}

func NewReader(scheme *runtime.Scheme) Reader {
	return newReader(scheme)
}

func newReader(scheme *runtime.Scheme) *reader {
	r := &reader{
		scheme:  scheme,
		decoder: serializer.NewCodecFactory(scheme).UniversalDeserializer(),
	}
	return r
}

func NewReaderFromFile(fpath string, scheme *runtime.Scheme) (Reader, error) {
	r := newReader(scheme)
	return r, r.readFile(fpath)
}

type reader struct {
	scheme  *runtime.Scheme
	decoder runtime.Decoder
	objects []object
}

type object struct {
	client.Object
	raw gojson.RawMessage
}

func (r *reader) AddObject(obj client.Object) error {
	bs, err := gojson.Marshal(obj)
	if err != nil {
		return err
	}

	r.objects = append(r.objects, object{
		Object: obj,
		raw:    bs,
	})
	return nil
}

func (r *reader) Get(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	gvk, err := apiutil.GVKForObject(obj, r.scheme)
	if err != nil {
		return err
	}

	for _, o := range r.objects {
		oGVK := o.GetObjectKind().GroupVersionKind()
		if o.GetName() == key.Name && o.GetNamespace() == key.Namespace && gvk.GroupKind() == oGVK.GroupKind() {
			return r.scheme.Convert(o.Object, obj, nil)
		}
	}

	return kerrors.NewNotFound(schema.GroupResource{Group: obj.GetObjectKind().GroupVersionKind().Group}, obj.GetName())
}

func (r *reader) List(_ context.Context, list client.ObjectList, opts ...client.ListOption) error {
	gvk, err := apiutil.GVKForObject(list, r.scheme)
	if err != nil {
		return err
	}

	gvk.Kind = strings.TrimSuffix(gvk.Kind, "List")

	listOpts := client.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(&listOpts)
	}

	var objs []gojson.RawMessage
	for _, o := range r.objects {
		oGVK := o.GetObjectKind().GroupVersionKind()
		if (listOpts.Namespace == "" || listOpts.Namespace == o.GetNamespace()) && gvk.GroupKind() == oGVK.GroupKind() {
			objs = append(objs, o.raw)
		}
	}

	bs, err := gojson.Marshal(map[string]interface{}{
		"items": objs,
	})
	if err != nil {
		return err
	}

	if _, _, err := r.decoder.Decode(bs, nil, list); err != nil {
		return err
	}

	return nil
}

func (r *reader) readFile(path string) error {
	bs, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	os, err := r.getObjectsFromContent(bs)
	if err != nil {
		return err
	}

	r.objects = os
	return nil
}

func (r *reader) getObjectsFromContent(bs []byte) ([]object, error) {
	s := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		r.scheme,
		r.scheme,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
		},
	)

	fr := json.YAMLFramer.NewFrameReader(io.NopCloser(bytes.NewReader(bs)))

	var res []object
	buffer := make([]byte, 1024)
	for {
		var buf bytes.Buffer
		for {
			n, err := fr.Read(buffer)
			buf.Write(buffer[:n])
			if err == io.EOF {
				return res, nil
			} else if err == io.ErrShortBuffer {
				continue
			} else if err != nil {
				return nil, err
			}

			break
		}

		ro, _, err := s.Decode(buf.Bytes(), nil, nil)
		if err != nil {
			return nil, err
		}

		if col, ok := ro.(client.ObjectList); ok {
			list := col.(*v1.List)
			for _, i := range list.Items {
				ro, _, err := s.Decode(i.Raw, nil, nil)
				if err != nil {
					return nil, err
				}

				co, ok := ro.(client.Object)
				if !ok {
					return nil,
						errors.UnimplementedErrorf("unknown object resource type in file '%v'",
							ro.GetObjectKind().GroupVersionKind())
				}

				res = append(res, object{
					Object: co,
					raw:    i.Raw,
				})
			}
			continue
		}

		co, ok := ro.(client.Object)
		if !ok {
			return nil, errors.UnimplementedErrorf("unknown object resource type in file '%v'",
				ro.GetObjectKind().GroupVersionKind())
		}

		raw, err := yaml.YAMLToJSON(buf.Bytes())
		if err != nil {
			return nil, err
		}

		res = append(res, object{
			Object: co,
			raw:    raw,
		})
	}
}
