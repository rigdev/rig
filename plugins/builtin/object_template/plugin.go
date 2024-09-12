// +groupName=plugins.rig.dev -- Only used for config doc generation
package objecttemplate

import (
	"bytes"
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/pipeline"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const Name = "rigdev.object_template"

// Configuration for the object_template plugin
// +kubebuilder:object:root=true
type Config struct {
	// The yaml to apply to the object. The yaml can be templated.
	Object string `json:"object,omitempty"`
	// Group to match, for which objects to apply the patch to.
	Group string `json:"group,omitempty"`
	// Kind to match, for which objects to apply the patch to.
	Kind string `json:"kind,omitempty"`
	// Name of the object to match. Default to Capsule-name. If '*' will execute the object template
	// on all objects of the given group and kind.
	Name string `json:"name,omitempty"`
}

type Plugin struct {
	plugin.NoWatchObjectStatus

	configBytes []byte
}

func (p *Plugin) ComputeConfig(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) (string, error) {
	return plugin.ParseCapsuleTemplatedConfigToString[Config](p.configBytes, req)
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	config, err := plugin.ParseCapsuleTemplatedConfig[Config](p.configBytes, req)
	if err != nil {
		return err
	}

	gk := schema.GroupVersionKind{
		Group: config.Group,
		Kind:  config.Kind,
	}

	var objects []client.Object
	if config.Name == "*" {
		objects, err = req.ListNew(gk)
		if err != nil {
			return err
		}
	} else {
		name := config.Name
		if name == "" {
			name = req.Capsule().Name
		}
		currentObject, err := req.GetNew(gk, name)
		if errors.IsNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}
		objects = append(objects, currentObject)
	}

	inputs, err := makeObjInputs(req, p.configBytes, objects)
	if err != nil {
		return err
	}

	for _, inp := range inputs {
		patchBytes, err := yaml.YAMLToJSON([]byte(inp.config.Object))
		if err != nil {
			return err
		}
		var currentBytes bytes.Buffer
		serializer := obj.NewSerializer(req.Scheme())
		if err := serializer.Encode(inp.obj, &currentBytes); err != nil {
			return err
		}

		modifiedBytes, err := strategicpatch.StrategicMergePatch(currentBytes.Bytes(), patchBytes, inp.obj)
		if err != nil {
			return err
		}

		if err := obj.DecodeInto(modifiedBytes, inp.obj, req.Scheme()); err != nil {
			return err
		}

		if err := req.Set(inp.obj); err != nil {
			return err
		}
	}

	return nil
}

type objInput struct {
	obj    client.Object
	config Config
}

func makeObjInputs(req pipeline.CapsuleRequest, originalConfigBytes []byte, objects []client.Object) ([]objInput, error) {
	var res []objInput
	for _, obj := range objects {
		config, err := plugin.ParseTemplatedConfig[Config](originalConfigBytes, req, plugin.CapsuleStep, func(_ Config, _ pipeline.CapsuleRequest) (map[string]any, error) {
			return map[string]any{
				"current": obj,
			}, nil
		})
		if err != nil {
			return nil, err
		}
		res = append(res, objInput{
			obj:    obj,
			config: config,
		})
	}

	return res, nil
}
