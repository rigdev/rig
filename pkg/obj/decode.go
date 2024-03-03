package obj

import (
	"bufio"
	"bytes"

	"github.com/rigdev/rig/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DecodeInto(bs []byte, into runtime.Object, scheme *runtime.Scheme) error {
	codecs := serializer.NewCodecFactory(scheme)

	info, _ := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)

	if _, _, err := info.Serializer.Decode(bs, nil, into); err != nil {
		return err
	}

	return nil
}

func Decode(bs []byte, out interface{}) error {
	r := yaml.NewYAMLToJSONDecoder(bufio.NewReader(bytes.NewReader(bs)))
	if err := r.Decode(out); err != nil {
		return errors.InvalidArgumentErrorf("bad yaml input: %v", err)
	}

	return nil
}

func DecodeAny(bs []byte, scheme *runtime.Scheme) (client.Object, error) {
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

	return ro.(client.Object), nil
}
