package main

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	Container *corev1.Container `json:"container"`
}

type initContainerPlugin struct {
	config Config
}

func (p *initContainerPlugin) Initialize(req plugin.InitializeRequest) error {
	return plugin.LoadYAMLConfig(req.Config, &p.config)
}

func (p *initContainerPlugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	c := *p.config.Container.DeepCopy()
	deployment.Spec.Template.Spec.InitContainers = append(deployment.Spec.Template.Spec.InitContainers, c)

	return req.Set(deployment)
}

func main() {
	plugin.StartPlugin("rigdev.init_container", &initContainerPlugin{})
}
