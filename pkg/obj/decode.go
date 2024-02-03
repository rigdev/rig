package obj

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func DecodeInto(bs []byte, into runtime.Object, scheme *runtime.Scheme) error {
	codecs := serializer.NewCodecFactory(scheme)

	info, _ := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)

	if _, _, err := info.Serializer.Decode(bs, nil, into); err != nil {
		return err
	}

	return nil
}
