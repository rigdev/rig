package plugin

import (
	"bytes"
	"html/template"

	"github.com/mitchellh/mapstructure"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/pipeline"
)

type ParseStep[T any] func(config T, req pipeline.CapsuleRequest) (string, any, error)

// ParseCapsuleTemplatedConfig parses the given data as a Go template with
// the capsule as a templating context under '.capsule'
// It then JSON/YAML decodes the resulting bytes into an instance of T
func ParseCapsuleTemplatedConfig[T any](data []byte, req pipeline.CapsuleRequest) (T, error) {
	return ParseTemplatedConfig(data, req, CapsuleStep[T])
}

// Using this, we parse the config at every execution of the plugin.
// If we get performance issues due to that we can try and optimize that.
func ParseTemplatedConfig[T any](data []byte, req pipeline.CapsuleRequest, steps ...ParseStep[T]) (T, error) {
	var config, empty T

	values := map[string]any{}
	for _, step := range steps {
		name, obj, err := step(config, req)
		if err != nil {
			return empty, err
		}

		result := map[string]interface{}{}
		d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &result})
		if err != nil {
			return empty, err
		}

		if err := d.Decode(obj); err != nil {
			return empty, err
		}
		values[name] = result
		t, err := template.New("config").Parse(string(data))
		if err != nil {
			return empty, err
		}
		var b bytes.Buffer
		if err := t.Execute(&b, values); err != nil {
			return empty, err
		}
		if err := LoadYAMLConfig(b.Bytes(), &config); err != nil {
			return empty, err
		}
	}

	return config, nil
}

func CapsuleStep[T any](_ T, req pipeline.CapsuleRequest) (string, any, error) {
	return "capsule", req.Capsule(), nil
}

func LoadYAMLConfig(data []byte, out any) error {
	return obj.Decode(data, out)
}
