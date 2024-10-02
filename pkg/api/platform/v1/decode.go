package v1

import (
	"encoding/json"
	"fmt"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/pkg/obj"
	"google.golang.org/protobuf/encoding/protojson"
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
	if err := YAMLToProto(bytes, spec, CapsuleKind); err != nil {
		return nil, err
	}
	return spec, nil
}

func CapsuleSetYAMLToProto(bytes []byte) (*platformv1.CapsuleSet, error) {
	spec := &platformv1.CapsuleSet{}
	if err := YAMLToProto(bytes, spec, CapsuleSetKind); err != nil {
		return nil, err
	}
	return spec, nil
}

func YAMLToProto[T interface{ GetKind() string }](bs []byte, o T, expectedKind string) error {
	if err := yaml.Unmarshal(bs, o, yaml.DisallowUnknownFields); err != nil {
		return fmt.Errorf("unmarshal: %s", err)
	}
	if o.GetKind() != "" && o.GetKind() != expectedKind {
		return fmt.Errorf("kind was %s, not the expected %s", o.GetKind(), expectedKind)
	}
	return nil
}

// ProtoToYAML converts a proto message to YAML in such a way that it empty
// structs, maps and lists are not included in the YAML. E.g. it will produce
//
// field1: value1
// field2: value2
//
// and not
//
// field1: value1
// field2: value2
// someList: []
// someObj:
//
//	child1: {}
//	child2: []
func ProtoToYAML(m proto.Message) (string, error) {
	m = proto.Clone(m)
	cleanProto(m)
	jsonString := protojson.Format(m)
	data, err := yaml.JSONToYAML([]byte(jsonString))
	if err != nil {
		return "", err
	}
	return string(data), nil
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
	InitialiseProto(res)
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
	InitialiseProto(res)
	return res
}

func InitialiseProto(msg proto.Message) {
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
			InitialiseProto(reflectMsg.Get(field).Message().Interface())
		}
	}
}

func DefaultCapsuleSpec() *platformv1.CapsuleSpec {
	spec := &platformv1.CapsuleSpec{}
	InitialiseProto(spec)
	spec.Scale.Horizontal.Instances.Min = 1
	spec.Scale.Vertical.Cpu.Request = "0.2"
	spec.Scale.Vertical.Memory.Request = "256Mi"
	return spec
}

func cleanProto(msg proto.Message) {
	if msg == nil {
		return
	}
	cleanReflectMessage(msg.ProtoReflect())
}

func cleanReflectMessage(msg protoreflect.Message) {
	if msg == nil {
		return
	}

	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		value := msg.Get(field)
		if field.IsMap() {
			m := value.Map()
			if m.Len() == 0 {
				msg.Clear(field)
			} else {
				cleanReflectMap(m, field.MapValue())
			}
		} else if field.IsList() {
			list := value.List()
			if list.Len() == 0 {
				msg.Clear(field)
			} else {
				cleanReflectList(list)
			}
		} else if field.Kind() == protoreflect.MessageKind {
			cleanReflectMessage(value.Message())
			if isEmpty(value.Message()) {
				msg.Clear(field)
			}
		}
	}
}

func cleanReflectList(list protoreflect.List) {
	if list.Len() == 0 {
		return
	}
	// A proto list either contains primitives or messages
	// There is unfortunately no way of directly getting the type/kind of list elements
	// so this hack must do.
	v := list.Get(0)
	if _, ok := v.Interface().(protoreflect.Message); ok {
		for idx := 0; idx < list.Len(); idx++ {
			cleanReflectMessage(list.Get(idx).Message())
		}
	}
}

func cleanReflectMap(m protoreflect.Map, valueDescriptor protoreflect.FieldDescriptor) {
	if m.Len() == 0 {
		return
	}
	// A map value is either a primitive or a message
	// We are only interested in messages
	m.Range(func(_ protoreflect.MapKey, v protoreflect.Value) bool {
		if valueDescriptor.Kind() == protoreflect.MessageKind {
			cleanReflectMessage(v.Message())
			return true
		}
		return false
	})
}

func isEmpty(msg protoreflect.Message) bool {
	if msg == nil {
		return true
	}
	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if !msg.Has(field) {
			continue
		}

		value := msg.Get(field)
		if field.IsMap() {
			if value.Map().Len() > 0 {
				return false
			}
		} else if field.IsList() {
			if value.List().Len() > 0 {
				return false
			}

		} else if field.Kind() == protoreflect.MessageKind {
			if value.Message().IsValid() && !isEmpty(value.Message()) {
				return false
			}
		} else {
			switch field.Kind() {
			case protoreflect.BoolKind:
				if value.Bool() {
					return false
				}
			case protoreflect.DoubleKind:
				fallthrough
			case protoreflect.FloatKind:
				if value.Float() != 0 {
					return false
				}
			case protoreflect.Int32Kind:
				fallthrough
			case protoreflect.Int64Kind:
				if value.Int() != 0 {
					return false
				}
			case protoreflect.StringKind:
				if value.String() != "" {
					return false
				}
			case protoreflect.Uint32Kind:
				fallthrough
			case protoreflect.Uint64Kind:
				if value.Uint() != 0 {
					return false
				}
			case protoreflect.BytesKind:
				if len(value.Bytes()) > 0 {
					return false
				}
			default:
				return false
			}
		}
	}

	return true
}
