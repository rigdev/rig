package main

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
)

type Config struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
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

	templateContext := plugin.NewTemplateContext()
	if err := templateContext.AddData("capsule", req.Capsule()); err != nil {
		return err
	}
	if err := templateContext.AddData("current", object); err != nil {
		return err
	}

	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	if err := handleMap(annotations, p.config.Annotations, templateContext); err != nil {
		return err
	}
	object.SetAnnotations(annotations)

	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	if err := handleMap(labels, p.config.Labels, templateContext); err != nil {
		return err
	}
	object.SetLabels(labels)

	return req.Set(object)
}

func handleMap(values map[string]string, updates map[string]string, templateContext *plugin.TemplateContext) error {
	for k, v := range updates {
		if v == "" {
			delete(values, k)
			continue
		}

		s, err := templateContext.Parse(v)
		if err != nil {
			return err
		}
		values[k] = s
	}
	return nil
}

func main() {
	plugin.StartPlugin("annotations", &annotationsPlugin{})
}
