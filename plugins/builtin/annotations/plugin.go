// +groupName=plugins.rig.dev -- Only used for config doc generation
package annotations

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/pipeline"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const Name = "rigdev.annotations"

// Configuration for the annotations plugin
// +kubebuilder:object:root=true
type Config struct {
	// Annotations are the annotations to insert into the object
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels are the labels to insert into the object
	Labels map[string]string `json:"labels,omitempty"`
	// Group to match, for which objects to apply the patch to.
	Group string `json:"group,omitempty"`
	// Kind to match, for which objects to apply the patch to.
	Kind string `json:"kind,omitempty"`
	// Name of the object to match. Defaults to Capsule-name.
	Name string `json:"name,omitempty"`
}

type Plugin struct {
	plugin.NoWatchObjectStatus

	configBytes []byte
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	config, err := plugin.ParseTemplatedConfig[Config](p.configBytes, req, plugin.CapsuleStep[Config])
	if err != nil {
		return err
	}

	name := config.Name
	if name == "" {
		name = req.Capsule().Name
	}

	object, err := req.GetNew(schema.GroupVersionKind{Group: config.Group, Kind: config.Kind}, name)
	if err != nil {
		return err
	}

	annotations := handleMap(object.GetAnnotations(), config.Annotations)
	object.SetAnnotations(annotations)

	labels := handleMap(object.GetLabels(), config.Labels)
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
