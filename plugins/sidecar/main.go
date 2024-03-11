// +groupName=plugins.rig.dev -- Only used for config doc generation
package main

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Configuration for the sidecar plugin
// +kubebuilder:object:root=true
type Config struct {
	// Container is the configuration of the sidecar injected into the deployment
	Container *corev1.Container `json:"container"`
}

type sidecar struct {
	configBytes []byte
}

func (s *sidecar) Initialize(req plugin.InitializeRequest) error {
	s.configBytes = req.Config
	return nil
}

func (s *sidecar) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	config, err := plugin.ParseTemplatedConfig[Config](s.configBytes, req, plugin.CapsuleStep[Config])
	if err != nil {
		return err
	}

	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	c := *config.Container.DeepCopy()
	c.RestartPolicy = ptr.New(corev1.ContainerRestartPolicyAlways)
	deployment.Spec.Template.Spec.InitContainers = append(deployment.Spec.Template.Spec.InitContainers, c)

	return req.Set(deployment)
}

func main() {
	plugin.StartPlugin("rigdev.sidecar", &sidecar{})
}
