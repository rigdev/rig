package base

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	RFC3339NanoFixed  = "2006-01-02T15:04:05.000000000Z07:00"
	RFC3339MilliFixed = "2006-01-02T15:04:05.000Z07:00"
)

func cmdPathContainsUsePrefix(cmd *cobra.Command, use string) bool {
	for cmd := cmd; cmd != nil; cmd = cmd.Parent() {
		if strings.HasPrefix(cmd.Use, use) {
			return true
		}
	}
	return false
}

func skipChecks(cmd *cobra.Command) bool {
	return cmdPathContainsUsePrefix(cmd, "completion") || cmdPathContainsUsePrefix(cmd, "help ")
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

func FormatPrint(v any) error {
	s, err := Format(v, Flags.OutputType)
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}
