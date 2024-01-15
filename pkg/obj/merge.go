package obj

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

type Merger interface {
	Merge(src, dst runtime.Object) (runtime.Object, error)
}

func NewMerger(scheme *runtime.Scheme) (Merger, error) {

	codecs := serializer.NewCodecFactory(scheme)

	info, _ := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)

	return &merger{
		serializer: info.Serializer,
	}, nil
}

type merger struct {
	serializer runtime.Serializer
}

func (m *merger) Merge(src, dst runtime.Object) (runtime.Object, error) {
	var srcB bytes.Buffer
	if err := m.serializer.Encode(src, &srcB); err != nil {
		return nil, fmt.Errorf("could not encode source obj: %w", err)
	}

	var dstB bytes.Buffer
	if err := m.serializer.Encode(dst, &dstB); err != nil {
		return nil, fmt.Errorf("could not encode destination obj: %w", err)
	}

	srcRN, err := yaml.Parse(srcB.String())
	if err != nil {
		return nil, fmt.Errorf("could not parse source yaml: %w", err)
	}

	dstRN, err := yaml.Parse(dstB.String())
	if err != nil {
		return nil, fmt.Errorf("could not parse destination yaml: %w", err)
	}

	resRN, err := merge2.Merge(srcRN, dstRN, yaml.MergeOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not merge documents: %w", err)
	}

	res, err := resRN.String()
	if err != nil {
		return nil, fmt.Errorf("could not reserialize merged document: %w", err)
	}

	obj, _, err := m.serializer.Decode([]byte(res), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decode merged document: %w", err)
	}

	return obj, nil
}
