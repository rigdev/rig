package obj

import (
	"bytes"
	"fmt"

	"github.com/rigdev/rig/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deprecaaated: use DecodeIntoT instead
func DecodeInto(bs []byte, into runtime.Object, scheme *runtime.Scheme) error {
	codecs := serializer.NewCodecFactory(scheme)

	info, _ := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)

	if _, _, err := info.Serializer.Decode(bs, nil, into); err != nil {
		return err
	}

	return nil
}

func DecodeIntoT[T runtime.Object](bs []byte, into T, scheme *runtime.Scheme) (T, error) {
	codecs := serializer.NewCodecFactory(scheme)

	info, _ := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)

	var empty T
	out, _, err := info.Serializer.Decode(bs, nil, into)
	if err != nil {
		return empty, err
	}

	t, ok := out.(T)
	if !ok {
		return empty, fmt.Errorf("decoded object had unexpected type %T", out)
	}

	return t, nil
}

func Decode(bs []byte, out any) error {
	r := yaml.NewYAMLToJSONDecoder(bytes.NewReader(bs))
	if err := r.Decode(out); err != nil {
		return errors.InvalidArgumentErrorf("bad yaml input: %v", err)
	}

	return nil
}

func DecodeAnyRuntime(bs []byte, scheme *runtime.Scheme) (runtime.Object, error) {
	s := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme,
		scheme,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
		},
	)

	ro, _, err := s.Decode(bs, nil, nil)
	if err != nil {
		return nil, err
	}
	return ro, nil
}

func DecodeAny(bs []byte, scheme *runtime.Scheme) (client.Object, error) {
	ro, err := DecodeAnyRuntime(bs, scheme)
	if err != nil {
		return nil, err
	}
	return ro.(client.Object), nil
}
