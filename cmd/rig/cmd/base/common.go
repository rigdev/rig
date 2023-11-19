package base

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
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

func Format(v any, outputType OutputType) (string, error) {
	switch outputType {
	case OutputTypeJSON:
		if v, ok := v.(protoreflect.ProtoMessage); ok {
			return protojson.Format(v), nil
		}
		bytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	case OutputTypeYAML:
		bytes, err := yaml.Marshal(v)
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
