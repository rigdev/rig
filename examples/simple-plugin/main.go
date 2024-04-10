package main

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/mod"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	appsv1 "k8s.io/api/apps/v1"
)

// This is an example of a minimal mod with some configuration.
// The mod adds a single label to the Deployment of the capsule.

// Config defines the configuration for the mod
type Config struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Mod implements the functionality
type Mod struct {
	configBytes []byte
}

func (m *Mod) Initialize(req mod.InitializeRequest) error {
	m.configBytes = req.Config
	return nil
}

func (m *Mod) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	config, err := mod.ParseTemplatedConfig[Config](m.configBytes, req, mod.CapsuleStep)
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
	mod.StartMod("myorg.simple", &Mod{})
}
