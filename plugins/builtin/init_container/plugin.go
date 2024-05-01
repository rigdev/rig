// +groupName=plugins.rig.dev -- Only used for config doc generation
package initcontainer

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/pipeline"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const Name = "rigdev.init_container"

// Configuration for the init_container plugin
// +kubebuilder:object:root=true
type Config struct {
	// Container holds the configuration for the init container
	Container *corev1.Container `json:"container"`
}

type Plugin struct {
	plugin.NoWatchObjectStatus

	configBytes []byte
	// config Config
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

	if config.Container == nil {
		return fmt.Errorf("invalid configuration, no `container` was specified")
	}

	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	c := *config.Container.DeepCopy()
	deployment.Spec.Template.Spec.InitContainers = append(deployment.Spec.Template.Spec.InitContainers, c)

	return req.Set(deployment)
}
