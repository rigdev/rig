package plugin

import (
	"bytes"
	"encoding/json"
	"html/template"

	"github.com/mitchellh/mapstructure"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/pipeline"
	"sigs.k8s.io/yaml"
)

type ParseStep[T any] func(config T, req pipeline.CapsuleRequest) (map[string]any, error)

// ParseCapsuleTemplatedConfig parses the given data as a Go template with
// the capsule as a templating context under '.capsule'
// It then JSON/YAML decodes the resulting bytes into an instance of T
func ParseCapsuleTemplatedConfig[T any](data []byte, req pipeline.CapsuleRequest) (T, error) {
	return ParseTemplatedConfig(data, req, CapsuleStep[T])
}

// Using this, we parse the config at every execution of the plugin.
// If we get performance issues due to that we can try and optimize that.
func ParseTemplatedConfig[T any](data []byte, req pipeline.CapsuleRequest, steps ...ParseStep[T]) (T, error) {
	if len(data) == 0 {
		data = []byte("{}")
	}
	var config, empty T

	values := map[string]any{}
	for _, step := range steps {
		m, err := step(config, req)
		if err != nil {
			return empty, err
		}

		for k, v := range m {
			result := map[string]any{}
			d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &result})
			if err != nil {
				return empty, err
			}
			if err := d.Decode(v); err != nil {
				return empty, err
			}
			values[k] = result
		}

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

func CapsuleStep[T any](_ T, req pipeline.CapsuleRequest) (map[string]any, error) {
	c := req.Capsule()
	extensions := map[string]any{}
	for k, v := range c.Spec.Extensions {
		vv := map[string]any{}
		if err := json.Unmarshal(v, &vv); err != nil {
			return nil, err
		}
		extensions[k] = vv
	}
	// Deep-copy before we manipulate the structure.
	c = c.DeepCopy()
	c.Spec.Extensions = nil
	return map[string]any{
		"capsule":           c,
		"capsuleExtensions": extensions,
	}, nil
}

func LoadYAMLConfig(data []byte, out any) error {
	return obj.Decode(data, out)
}

func ParseCapsuleTemplatedConfigToString[T any](data []byte, req pipeline.CapsuleRequest) (string, error) {
	obj, err := ParseCapsuleTemplatedConfig[T](data, req)
	if err != nil {
		return "", err
	}
	bs, err := yaml.Marshal(obj)
	return string(bs), err
}
