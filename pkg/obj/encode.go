package obj

import (
	"bytes"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/yaml"
)

func Encode(obj runtime.Object, scheme *runtime.Scheme) ([]byte, error) {
	codecs := serializer.NewCodecFactory(scheme)

	info, _ := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)

	var buffer bytes.Buffer
	if err := info.Serializer.Encode(obj, &buffer); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func EncodeAny(obj any) ([]byte, error) {
	return yaml.Marshal(obj)
}
