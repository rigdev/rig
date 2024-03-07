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

type pluginParent struct {
	configBytes []byte
}

func (p *pluginParent) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *pluginParent) Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error {
	config, err := plugin.ParseTemplatedConfig[Config](p.configBytes, req, plugin.CapsuleStep[Config])
	if err != nil {
		return err
	}
	pp := &pluginImpl{
		config: config,
	}
	return pp.run(ctx, req, logger)
}

type pluginImpl struct {
	config Config
}

func (p *pluginImpl) run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	name := p.config.Name
	if name == "" {
		name = req.Capsule().Name
	}

	object, err := plugin.GetNew(p.config.Group, p.config.Kind, name, req)
	if err != nil {
		return err
	}

	annotations := handleMap(object.GetAnnotations(), p.config.Annotations)
	object.SetAnnotations(annotations)

	labels := handleMap(object.GetLabels(), p.config.Labels)
	object.SetLabels(labels)

	return req.Set(object)
}

func handleMap(values map[string]string, updates map[string]string) map[string]string {
	if values == nil {
		values = map[string]string{}
	}
	for k, v := range updates {
		if v == "" {
			delete(values, k)
			continue
		}
		values[k] = v
	}
	return values
}

func main() {
	plugin.StartPlugin("rigdev.annotations", &pluginParent{})
}
