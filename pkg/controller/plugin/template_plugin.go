package plugin

import (
	"bytes"
	"context"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/obj"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type TemplatePlugin struct {
	Template string `json:"template"`
	Group    string `json:"group"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
}

func NewTemplatePlugin(config map[string]string) (Plugin, error) {
	p := &TemplatePlugin{}
	return p, mapstructure.Decode(config, p)
}

func (s *TemplatePlugin) Run(_ context.Context, req pipeline.Request) error {
	gvk, err := pipeline.LookupGVK(schema.GroupKind{Group: s.Group, Kind: s.Kind})
	if err != nil {
		return err
	}

	name := s.Name
	if name == "" {
		name = req.Capsule().Name
	}

	key := req.NamedObjectKey(name, gvk)

	object := req.GetNew(key)
	if object == nil {
		return nil
	}

	t, err := template.New("plugin").Parse(s.Template)
	if err != nil {
		return err
	}

	input := map[string]interface{}{
		"capsule": req.Capsule(),
		"current": object,
	}

	values := map[string]interface{}{}
	for name, in := range input {
		result := map[string]interface{}{}
		d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &result})
		if err != nil {
			return err
		}

		if err := d.Decode(in); err != nil {
			return err
		}

		values[name] = result
	}

	var out bytes.Buffer
	if err := t.Execute(&out, values); err != nil {
		return err
	}

	new, err := req.Scheme().New(gvk)
	if err != nil {
		return err
	}

	if err := obj.DecodeInto(out.Bytes(), new, req.Scheme()); err != nil {
		return err
	}

	merge := obj.NewMerger(req.Scheme())
	if err := merge.Merge(new, object); err != nil {
		return err
	}

	req.Set(key, object)
	return nil
}
