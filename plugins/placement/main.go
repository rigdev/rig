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
	NodeSelector      map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations       []corev1.Toleration `json:"tolerations,omitempty"`
	RequireStepName   bool                `json:"requireStepName,omitempty"`
	RequirePluginName bool                `json:"requirePluginName,omitempty"`
}

type placement struct {
	configBytes []byte

	config     Config
	stepName   string
	pluginName string
}

const (
	StepNameAnnotation   = "rigdev.placement/step_name"
	PluginNameAnnotation = "rigdev.placement/plugin_name"
)

func (p *placement) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	p.stepName = req.StepName
	p.pluginName = req.PluginName
	return nil
}

func (p *placement) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
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
	deployment.Spec.Template.Spec.Tolerations = append(deployment.Spec.Template.Spec.Tolerations, p.config.Tolerations...)

	return req.Set(deployment)
}

func (p *placement) shouldRun(req pipeline.CapsuleRequest) bool {
	capsule := req.Capsule()
	if p.config.RequireStepName {
		if p.stepName == "" {
			return false
		}
		if v := capsule.Annotations[StepNameAnnotation]; v != p.stepName {
			return false
		}
	}
	if p.config.RequirePluginName {
		if p.pluginName == "" {
			return false
		}
		if v := capsule.Annotations[PluginNameAnnotation]; v != p.pluginName {
			return false
		}
	}
	return true
}

func main() {
	plugin.StartPlugin("rigdev.placement", &placement{})
}
