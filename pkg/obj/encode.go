package obj

import (
	"bytes"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func Dump(obj client.Object, scheme *runtime.Scheme) {
	codecs := serializer.NewCodecFactory(scheme)

	info, _ := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)

	var buffer bytes.Buffer
	if err := info.Serializer.Encode(obj, &buffer); err != nil {
		panic(err)
	}

	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(os.Stderr, "*************** OBJ DUMP START ***************")
	fmt.Fprintln(os.Stderr, gvks[0])
	fmt.Fprintf(os.Stderr, "%s/%s\n", obj.GetNamespace(), obj.GetName())
	fmt.Fprintln(os.Stderr, buffer.String())
	fmt.Fprintln(os.Stderr, "***************  OBJ DUMP END  ***************")
}
