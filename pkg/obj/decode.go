package obj

import (
	"bufio"
	"bytes"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func DecodeInto(bs []byte, into runtime.Object, scheme *runtime.Scheme) error {
	codecs := serializer.NewCodecFactory(scheme)

	info, _ := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)

	if _, _, err := info.Serializer.Decode(bs, nil, into); err != nil {
		return err
	}

	return nil
}

func DecodeYAML(bs []byte, out interface{}) error {
	r := yaml.NewYAMLToJSONDecoder(bufio.NewReader(bytes.NewReader(bs)))
	return r.Decode(out)
}
