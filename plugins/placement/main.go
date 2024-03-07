package main

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	NodeSelector    map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations     []corev1.Toleration `json:"tolerations,omitempty"`
	RequireStepID   bool                `json:"requireStepID,omitempty"`
	RequirePluginID bool                `json:"requirePluginID,omitempty"`
}

type placement struct {
	configBytes []byte

	config   Config
	stepID   string
	pluginID string
}

const (
	StepIDAnnotation   = "rigdev.placement/step_id"
	PluginIDAnnotation = "rigdev.placement/plugin_id"
)

func (p *placement) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	p.stepID = req.StepID
	p.pluginID = req.PluginID
	return nil
}

func (p *placement) Run(_ context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error {
	var err error
	p.config, err = plugin.ParseTemplatedConfig[Config](p.configBytes, req, plugin.CapsuleStep[Config])
	if err != nil {
		return err
	}

	if !p.shouldRun(req) {
		return nil
	}

	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); err != nil {
		return err
	}

	selector := deployment.Spec.Template.Spec.NodeSelector
	if selector == nil {
		selector = map[string]string{}
	}
	for k, v := range p.config.NodeSelector {
		selector[k] = v
	}
	deployment.Spec.Template.Spec.NodeSelector = selector

	for _, t := range p.config.Tolerations {
		deployment.Spec.Template.Spec.Tolerations = append(deployment.Spec.Template.Spec.Tolerations, t)
	}

	return req.Set(deployment)
}

func (p *placement) shouldRun(req pipeline.CapsuleRequest) bool {
	capsule := req.Capsule()
	if p.config.RequireStepID {
		if p.stepID == "" {
			return false
		}
		if v := capsule.Annotations[StepIDAnnotation]; v != p.stepID {
			return false
		}
	}
	if p.config.RequirePluginID {
		if p.pluginID == "" {
			return false
		}
		if v := capsule.Annotations[PluginIDAnnotation]; v != p.pluginID {
			return false
		}
	}
	return true
}

func main() {
	plugin.StartPlugin("rigdev.placement", &placement{})
}
