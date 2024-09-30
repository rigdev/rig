package v1

import (
	"encoding/json"
	"fmt"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/pkg/obj"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

func CapsuleSpecYAMLToProto(bytes []byte) (*platformv1.CapsuleSpec, error) {
	spec := &platformv1.CapsuleSpec{}
	if err := yaml.Unmarshal(bytes, spec, yaml.DisallowUnknownFields); err != nil {
		return nil, fmt.Errorf("unmarshal: %s", err)
	}
	return spec, nil
}

func CapsuleYAMLToProto(bytes []byte) (*platformv1.Capsule, error) {
	spec := &platformv1.Capsule{}
	if err := YAMLToSpecProto(bytes, spec, CapsuleKind); err != nil {
		return nil, err
	}
	return spec, nil
}

func CapsuleSetYAMLToProto(bytes []byte) (*platformv1.CapsuleSet, error) {
	spec := &platformv1.CapsuleSet{}
	if err := YAMLToSpecProto(bytes, spec, CapsuleSetKind); err != nil {
		return nil, err
	}
	return spec, nil
}

func YAMLToSpecProto[T interface{ GetKind() string }](bs []byte, o T, expectedKind string) error {
	if err := yaml.Unmarshal(bs, o, yaml.DisallowUnknownFields); err != nil {
		return fmt.Errorf("unmarshal: %s", err)
	}
	if o.GetKind() != "" && o.GetKind() != expectedKind {
		return fmt.Errorf("kind was %s, not the expected %s", o.GetKind(), expectedKind)
	}
	return nil
}

func CapsuleProtoToCRD(capsule *platformv1.Capsule, scheme *runtime.Scheme) (*Capsule, error) {
	// Use json instead of yaml to omit empty fields
	data, err := json.Marshal(capsule)
	if err != nil {
		return nil, err
	}

	res, err := obj.DecodeIntoT(data, &Capsule{}, scheme)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func CapsuleSetProtoToCRD(capsuleSet *platformv1.CapsuleSet, scheme *runtime.Scheme) (*CapsuleSet, error) {
	// Use json instead of yaml to omit empty fields
	data, err := json.Marshal(capsuleSet)
	if err != nil {
		return nil, err
	}

	res, err := obj.DecodeIntoT(data, &CapsuleSet{}, scheme)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func CapsuleSpecProtoToCRD(spec *platformv1.CapsuleSpec, scheme *runtime.Scheme) (CapsuleSpec, error) {
	c := &platformv1.Capsule{
		Spec:       spec,
		Kind:       CapsuleKind,
		ApiVersion: GroupVersion.String(),
	}
	crd, err := CapsuleProtoToCRD(c, scheme)
	if err != nil {
		return CapsuleSpec{}, err
	}
	return crd.Spec, nil
}

func CapsuleCRDToProto(capsule *Capsule, scheme *runtime.Scheme) (*platformv1.Capsule, error) {
	data, err := obj.Encode(capsule, scheme)
	if err != nil {
		return nil, err
	}
	return CapsuleYAMLToProto(data)
}

func CapsuleSetCRDToProto(capsule *CapsuleSet, scheme *runtime.Scheme) (*platformv1.CapsuleSet, error) {
	data, err := obj.Encode(capsule, scheme)
	if err != nil {
		return nil, err
	}
	return CapsuleSetYAMLToProto(data)
}

func CapsuleSpecCRDToProto(spec CapsuleSpec, scheme *runtime.Scheme) (*platformv1.CapsuleSpec, error) {
	c := &Capsule{
		TypeMeta: metav1.TypeMeta{
			Kind:       CapsuleKind,
			APIVersion: GroupVersion.String(),
		},
		Spec: spec,
	}
	data, err := obj.Encode(c, scheme)
	if err != nil {
		return nil, err
	}
	cp, err := CapsuleYAMLToProto(data)
	if err != nil {
		return nil, err
	}
	return cp.GetSpec(), nil
}

func NewCapsuleProto(projectID, environmentID, capsuleID string, spec *platformv1.CapsuleSpec) *platformv1.Capsule {
	res := &platformv1.Capsule{
		Kind:        CapsuleKind,
		ApiVersion:  GroupVersion.String(),
		Name:        capsuleID,
		Project:     projectID,
		Environment: environmentID,
		Spec:        spec,
	}
	initialise(res)
	return res
}

func NewCapsuleSetProto(projectID, capsuleID string, spec *platformv1.CapsuleSpec) *platformv1.CapsuleSet {
	res := &platformv1.CapsuleSet{
		Kind:       CapsuleSetKind,
		ApiVersion: GroupVersion.String(),
		Name:       capsuleID,
		Project:    projectID,
		Spec:       spec,
	}
	initialise(res)
	return res
}

func initialise(msg proto.Message) {
	reflectMsg := msg.ProtoReflect()
	fields := reflectMsg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if field.IsMap() {
			if !reflectMsg.Has(field) {
				reflectMsg.Set(field, reflectMsg.NewField(field))
			}
		} else if field.IsList() {
		} else if field.Kind() == protoreflect.MessageKind {
			if !reflectMsg.Has(field) {
				reflectMsg.Set(field, reflectMsg.NewField(field))
			}
			initialise(reflectMsg.Get(field).Message().Interface())
		}

	}
}
