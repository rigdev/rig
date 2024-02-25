package main

import (
	"bytes"
	"context"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/mapstructure"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	// The yaml to apply to the object. The yaml can be templated.
	Object string `json:"object,omitempty"`
	// Group to match, for which objects to apply the patch to.
	Group string `json:"group,omitempty"`
	// Kind to match, for which objects to apply the patch to.
	Kind string `json:"kind,omitempty"`
	// Name of the object to match. Default to Capsule-name.
	Name string `json:"name,omitempty"`
}

type objectTemplatePlugin struct {
	config Config
}

func (p *objectTemplatePlugin) LoadConfig(data []byte) error {
	return plugin.LoadYAMLConfig(data, &p.config)
}

func (p *objectTemplatePlugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	gvk, err := pipeline.LookupGVK(schema.GroupKind{Group: p.config.Group, Kind: p.config.Kind})
	if err != nil {
		return err
	}

	co, err := req.Scheme().New(gvk)
	if err != nil {
		return err
	}

	cc := co.(client.Object)
	cc.SetName(p.config.Name)

	if err := req.GetNew(cc); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	t, err := template.New("plugin").Parse(p.config.Object)
	if err != nil {
		return err
	}

	input := map[string]interface{}{
		"capsule": req.Capsule(),
		"current": cc,
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

	patch, err := req.Scheme().New(gvk)
	if err != nil {
		return err
	}

	if err := obj.DecodeInto(out.Bytes(), patch, req.Scheme()); err != nil {
		return err
	}

	merge := obj.NewMerger(req.Scheme())
	if err := merge.Merge(patch, cc); err != nil {
		return err
	}

	return req.Set(cc)
}

func main() {
	plugin.StartPlugin("objectTemplate", &objectTemplatePlugin{})
}
