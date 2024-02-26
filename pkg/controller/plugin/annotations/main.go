package main

import (
	"bytes"
	"context"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
)

type Config struct {
	Annotations map[string]string
	// Group to match, for which objects to apply the patch to.
	Group string `json:"group,omitempty"`
	// Kind to match, for which objects to apply the patch to.
	Kind string `json:"kind,omitempty"`
	// Name of the object to match. Default to Capsule-name.
	Name string `json:"name,omitempty"`
}

type annotationsPlugin struct {
	config Config
}

func (p *annotationsPlugin) LoadConfig(data []byte) error {
	return plugin.LoadYAMLConfig(data, &p.config)
}

func (p *annotationsPlugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	name := p.config.Name
	if name == "" {
		name = req.Capsule().Name
	}
	object, err := plugin.GetNew(p.config.Group, p.config.Kind, name, req)
	if err != nil {
		return err
	}

	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	values, err := plugin.TemplateDataUsingJSONTags(map[string]interface{}{
		"capsule": req.Capsule(),
		"current": object,
	})
	if err != nil {
		return err
	}

	for k, v := range p.config.Annotations {
		if v == "" {
			delete(annotations, k)
			continue
		}

		t, err := template.New("annotation").Parse(v)
		if err != nil {
			return err
		}
		var buffer bytes.Buffer
		if err := t.Execute(&buffer, values); err != nil {
			return err
		}
		annotations[k] = buffer.String()
	}
	object.SetAnnotations(annotations)

	return req.Set(object)
}

func main() {
	plugin.StartPlugin("annotations", &annotationsPlugin{})
}
