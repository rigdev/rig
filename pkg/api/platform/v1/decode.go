package v1

import (
	"fmt"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"gopkg.in/yaml.v3"
)

func YAMLToCapsuleSpecProto(bytes []byte) (*platformv1.CapsuleSpec, error) {
	spec := &platformv1.CapsuleSpec{}
	if err := YAMLToSpecProto(bytes, spec, "CapsuleSpec"); err != nil {
		return nil, err
	}
	return spec, nil
}

func YAMLToCapsuleProto(bytes []byte) (*platformv1.Capsule, error) {
	spec := &platformv1.Capsule{}
	if err := YAMLToSpecProto(bytes, spec, "Capsule"); err != nil {
		return nil, err
	}
	return spec, nil
}

func YAMLToSpecProto[T interface{ GetKind() string }](bytes []byte, obj T, expectedKind string) error {
	if err := yaml.Unmarshal(bytes, obj); err != nil {
		return err
	}
	if obj.GetKind() != "" && obj.GetKind() != expectedKind {
		return fmt.Errorf("kind was %s, not the expected %s", obj.GetKind(), expectedKind)
	}
	return nil
}
