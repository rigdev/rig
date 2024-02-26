package main

import (
	"bytes"
	"context"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/yaml"
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
	name := p.config.Name
	if name == "" {
		name = req.Capsule().Name
	}
	currentObject, err := plugin.GetNew(p.config.Group, p.config.Kind, name, req)
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	values, err := plugin.TemplateDataUsingJSONTags(map[string]interface{}{
		"capsule": req.Capsule(),
		"current": currentObject,
	})
	if err != nil {
		return err
	}

	t, err := template.New("plugin").Parse(p.config.Object)
	if err != nil {
		return err
	}
	var patchBuffer bytes.Buffer
	if err := t.Execute(&patchBuffer, values); err != nil {
		return err
	}
	patchBytes, err := yaml.YAMLToJSON(patchBuffer.Bytes())
	if err != nil {
		return err
	}

	var currentBytes bytes.Buffer
	serializer := obj.NewSerializer(req.Scheme())
	if err := serializer.Encode(currentObject, &currentBytes); err != nil {
		return err
	}

	modifiedBytes, err := strategicpatch.StrategicMergePatch(currentBytes.Bytes(), patchBytes, currentObject)
	if err != nil {
		return err
	}

	if err := obj.DecodeInto(modifiedBytes, currentObject, req.Scheme()); err != nil {
		return err
	}

	return req.Set(currentObject)
}

func main() {
	plugin.StartPlugin("objectTemplate", &objectTemplatePlugin{})
}
