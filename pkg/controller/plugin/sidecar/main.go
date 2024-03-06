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

type Config struct {
	Container *corev1.Container `json:"container"`
}

type pluginParent struct {
	configBytes []byte
}

func (p *pluginParent) LoadConfig(data []byte) error {
	p.configBytes = data
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
	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	c := *p.config.Container.DeepCopy()
	c.RestartPolicy = ptr.New(corev1.ContainerRestartPolicyAlways)
	deployment.Spec.Template.Spec.InitContainers = append(deployment.Spec.Template.Spec.InitContainers, c)

	return req.Set(deployment)
}

func main() {
	plugin.StartPlugin("rigdev.sidecar", &pluginParent{})
}
