package obj

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

func NewSerializer(scheme *runtime.Scheme) runtime.Serializer {
	codecs := serializer.NewCodecFactory(scheme)
	info, _ := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeJSON)
	return info.Serializer
}

// Merge applies 'source' as a patch to 'dest', returning a new object containing the merge result.
// It requires that 'dest' has APIVersion and Kind set.
// This function is dangerous to use with objects derived from yaml/json when fields don't have omitEmpty
// Consider a field with no omitEmpty JSON tag. If this field is not set in 'source', if will be present as an empty
// value in the JSON of 'source' and then overwrite whatever the field had in 'dest'.
func Merge[T runtime.Object](patch runtime.Object, object runtime.Object, result T, serializer runtime.Serializer) (T, error) {
	var empty T

	var srcB bytes.Buffer
	if err := serializer.Encode(patch, &srcB); err != nil {
		return empty, fmt.Errorf("could not encode source obj: %w", err)
	}

	var dstB bytes.Buffer
	if err := serializer.Encode(object, &dstB); err != nil {
		return empty, fmt.Errorf("could not encode destination obj: %w", err)
	}

	out, err := strategicpatch.StrategicMergePatch(dstB.Bytes(), srcB.Bytes(), object)
	if err == nil {
		res, _, err := serializer.Decode(out, nil, result)
		if err != nil {
			return empty, err
		}
		return convert[T](res)
	}

	// Alternative solution:
	// out, err := jsonpatch.MergePatch(dstB.Bytes(), srcB.Bytes())
	// if err != nil {
	// 	fmt.Println("did not work")
	// 	return err
	// }

	// _, _, err = m.serializer.Decode(out, nil, dst)
	// if err != nil {
	// 	return fmt.Errorf("could not decode merged document: %w", err)
	// }

	// return nil

	srcRN, err := yaml.Parse(srcB.String())
	if err != nil {
		return empty, fmt.Errorf("could not parse source yaml: %w", err)
	}

	dstRN, err := yaml.Parse(dstB.String())
	if err != nil {
		return empty, fmt.Errorf("could not parse destination yaml: %w", err)
	}

	resRN, err := merge2.Merge(srcRN, dstRN, yaml.MergeOptions{})
	if err != nil {
		return empty, fmt.Errorf("could not merge documents: %w", err)
	}

	res, err := resRN.String()
	if err != nil {
		return empty, fmt.Errorf("could not reserialize merged document: %w", err)
	}

	output, _, err := serializer.Decode([]byte(res), nil, nil)
	if err != nil {
		return empty, fmt.Errorf("could not decode merged document: %w", err)
	}

	return convert[T](output)
}

func convert[T any](obj any) (T, error) {
	objT, ok := obj.(T)
	if !ok {
		var empty T
		return empty, fmt.Errorf("expected output to have type %T, it had type %T", empty, obj)
	}
	return objT, nil
}
