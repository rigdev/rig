package main

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/pipeline"
	appsv1 "k8s.io/api/apps/v1"
)

// This is an example of a minimal plugin with some configuration.
// The plugin adds a single label to the Deployment of the capsule.

// Config defines the configuration for the plugin
type Config struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Plugin implements the functionality
type Plugin struct {
	configBytes []byte
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	config, err := plugin.ParseTemplatedConfig[Config](p.configBytes, req, plugin.CapsuleStep)
	if err != nil {
		return err
	}

	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); err != nil {
		return err
	}

	if deployment.Labels == nil {
		deployment.Labels = map[string]string{}
	}

	deployment.Labels[config.Label] = config.Value

	return req.Set(deployment)
}

func main() {
	plugin.StartPlugin("myorg.simple", &Plugin{})
}
