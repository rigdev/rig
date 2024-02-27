package plugin

import (
	"bytes"
	"html/template"

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

type TemplateContext struct {
	values map[string]any
}

func NewTemplateContext() *TemplateContext {
	return &TemplateContext{
		values: map[string]any{},
	}
}

func (t *TemplateContext) Parse(s string) (string, error) {
	tt, err := template.New("value").Parse(s)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err := tt.Execute(&buffer, t.values); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (t *TemplateContext) AddData(name string, data any) error {
	result := map[string]interface{}{}
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &result})
	if err != nil {
		return err
	}

	if err := d.Decode(data); err != nil {
		return err
	}
	t.values[name] = result
	return nil
}
