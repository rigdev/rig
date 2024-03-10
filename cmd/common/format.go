package common

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ghodss/yaml"
	"github.com/rigdev/rig/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type OutputType string

const (
	OutputTypeJSON   OutputType = "json"
	OutputTypeYAML   OutputType = "yaml"
	OutputTypePretty OutputType = "pretty"
)

// String is used both by fmt.Print and by Cobra in help text
func (e *OutputType) String() string {
	return string(*e)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *OutputType) Set(v string) error {
	switch v {
	case string(OutputTypeJSON), string(OutputTypeYAML), string(OutputTypePretty):
		*e = OutputType(v)
		return nil
	default:
		return errors.New(`must be one of "json", "yaml", or "pretty"`)
	}
}

// Type is only used in help text
func (e *OutputType) Type() string {
	return `pretty,json,yaml`
}

// If 'v' is of type protoreflect.Message or a slice of protoreflect.Message
// we want to use protojson.Format otherwise we get wrong json names
// Unfortunately, protojson.Format does not accept a list of protoreflect.Message
// thus we do a little hacking around it
func formatJSON(v any) (string, error) {
	value := reflect.ValueOf(v)
	if value.Type().Kind() == reflect.Slice {
		allProto := true
		var protoSlices []protoreflect.ProtoMessage
		for i := 0; i < value.Len(); i++ {
			v := value.Index(i).Interface()
			if pm, ok := v.(protoreflect.ProtoMessage); ok {
				protoSlices = append(protoSlices, pm)
			} else {
				allProto = false
			}
		}
		if allProto {
			var slice []any
			for _, v := range protoSlices {
				s := protojson.Format(v)
				var obj any
				if err := json.Unmarshal([]byte(s), &obj); err != nil {
					return "", err
				}
				slice = append(slice, obj)
			}
			bs, err := json.MarshalIndent(slice, "", "  ")
			return string(bs), err
		}
		bs, err := json.MarshalIndent(v, "", "  ")
		return string(bs), err
	}

	if v, ok := v.(protoreflect.ProtoMessage); ok {
		return protojson.Format(v), nil
	}
	bs, err := json.MarshalIndent(v, "", "  ")
	return string(bs), err
}

func Format(v any, outputType OutputType) (string, error) {
	switch outputType {
	case OutputTypeJSON:
		return formatJSON(v)
	case OutputTypeYAML:
		// We need to first get the JSON string as protoreflect.Message
		// objects do not get correctly marshalled directly as yaml
		jsonString, err := formatJSON(v)
		if err != nil {
			return "", err
		}
		bytes, err := yaml.JSONToYAML([]byte(jsonString))
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	default:
		return "", fmt.Errorf("unexpected output type %v", outputType)
	}
}

func FormatPrint(v any, o OutputType) error {
	s, err := Format(v, o)
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}
