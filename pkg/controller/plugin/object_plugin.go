package plugin

import (
	"bytes"
	"context"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/obj"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ObjectPlugin struct {
	cfg *v1alpha1.ObjectPlugin
}

func NewObjectPlugin(cfg *v1alpha1.ObjectPlugin) Plugin {
	return &ObjectPlugin{
		cfg: cfg,
	}
}

func (s *ObjectPlugin) Run(_ context.Context, req pipeline.Request) error {
	gvk, err := pipeline.LookupGVK(schema.GroupKind{Group: s.cfg.Group, Kind: s.cfg.Kind})
	if err != nil {
		return err
	}

	name := s.cfg.Name
	if name == "" {
		name = req.Capsule().Name
	}

	key := req.NamedObjectKey(name, gvk)

	object := req.GetNew(key)
	if object == nil {
		return nil
	}

	t, err := template.New("plugin").Parse(s.cfg.Object)
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
