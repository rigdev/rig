package obj

import (
	"bytes"
	"encoding/json"
	"fmt"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	v1 "github.com/rigdev/rig/pkg/api/platform/v1"
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
func Merge[T runtime.Object](patch runtime.Object,
	object runtime.Object,
	result T,
	serializer runtime.Serializer) (T, error) {
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

// MergeProjectEnv merges a ProjEnvCapsuleBase into a CapsuleSpecExtension and returns a new object with the merged result
// It uses StrategicMergePatch (https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/)
func MergeProjectEnv(patch *platformv1.ProjEnvCapsuleBase, into *platformv1.CapsuleSpecExtension) (*platformv1.CapsuleSpecExtension, error) {
	return mergeCapsuleSpec(patch, into)
}

// MergeCapsuleSpecExtension merges a CapsuleSpecExtension into another CapsuleSpecExtension and returns a new object with the merged result
// It uses StrategicMergePatch (https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/)
func MergeCapsuleSpecExtensions(patch, into *platformv1.CapsuleSpecExtension) (*platformv1.CapsuleSpecExtension, error) {
	return mergeCapsuleSpec(patch, into)
}

func mergeCapsuleSpec(patch any, into *platformv1.CapsuleSpecExtension) (*platformv1.CapsuleSpecExtension, error) {
	// It would be possible to do much faster merging by manualling overwriting protobuf fields.
	// This is tedius to maintain so until it becomes an issue, we use json marshalling to leverage StrategicMergePatch
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	intoBytes, err := json.Marshal(into)
	if err != nil {
		return nil, err
	}

	outBytes, err := strategicpatch.StrategicMergePatch(intoBytes, patchBytes, &v1.CapsuleSpecExtension{})
	if err != nil {
		return nil, err
	}

	out := &platformv1.CapsuleSpecExtension{}
	if err := json.Unmarshal(outBytes, out); err != nil {
		return nil, err
	}
	out.Kind = into.GetKind()
	out.ApiVersion = into.GetApiVersion()

	return out, nil
}
