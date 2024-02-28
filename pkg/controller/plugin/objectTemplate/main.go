package main

import (
	"bytes"
	"context"

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

type pluginParent struct {
	configBytes []byte
}

func (p *pluginParent) LoadConfig(data []byte) error {
	p.configBytes = data
	return nil
}

func (p *pluginParent) Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error {
	config, err := plugin.ParseTemplatedConfig[Config](p.configBytes, req,
		plugin.CapsuleStep[Config],
		func(c Config, req pipeline.CapsuleRequest) (string, any, error) {
			name := c.Name
			if name == "" {
				name = req.Capsule().Name
			}
			currentObject, err := plugin.GetNew(c.Group, c.Kind, name, req)
			if err != nil {
				return "", nil, err
			}
			return "current", currentObject, nil
		},
	)
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
	currentObject, err := plugin.GetNew(p.config.Group, p.config.Kind, name, req)
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	patchBytes, err := yaml.YAMLToJSON([]byte(p.config.Object))
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
	plugin.StartPlugin("objectTemplate", &pluginParent{})
}
