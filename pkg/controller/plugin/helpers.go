package plugin

import (
	"github.com/mitchellh/mapstructure"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetNew(group, kind, name string, req pipeline.CapsuleRequest) (client.Object, error) {
	gvk, err := pipeline.LookupGVK(schema.GroupKind{Group: group, Kind: kind})
	if err != nil {
		return nil, err
	}

	co, err := req.Scheme().New(gvk)
	if err != nil {
		return nil, err
	}

	currentObject := co.(client.Object)
	currentObject.SetName(name)

	if err := req.GetNew(currentObject); err != nil {
		return nil, err
	}

	return currentObject, nil
}

func TemplateDataUsingJSONTags(input map[string]any) (map[string]any, error) {
	values := map[string]interface{}{}
	for name, in := range input {
		result := map[string]interface{}{}
		d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &result})
		if err != nil {
			return nil, err
		}

		if err := d.Decode(in); err != nil {
			return nil, err
		}

		values[name] = result
	}
	return values, nil
}
