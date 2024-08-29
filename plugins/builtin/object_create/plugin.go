// +groupName=plugins.rig.dev -- Only used for config doc generation
package objectcreate

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/pipeline"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const Name = "rigdev.object_create"

// Configuration for the object_create plugin
// +kubebuilder:object:root=true
type Config struct {
	// The yaml to apply as an object. The yaml can be templated.
	Object string `json:"object,omitempty"`
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
	config, err := plugin.ParseCapsuleTemplatedConfig[Config](p.configBytes, req)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if config.Object == "" {
		return nil
	}

	object := &unstructured.Unstructured{}
	if err := obj.DecodeInto([]byte(config.Object), object, req.Scheme()); err != nil {
		return err
	}

	if err := req.Set(object); err != nil {
		return err
	}

	return nil
}
