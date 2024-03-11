// +groupName=plugins.rig.dev -- Only used for config doc generation
package main

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Configuration for the placement plugin
// +kubebuilder:object:root=true
type Config struct {
	// Nodeselectors which will be inserted into the deployment's podSpec
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations which will be appended to the deployment's podSpec
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// True if a capsule needs a Tag annotation to be run
	RequireTag bool `json:"requireTag,omitempty"`
}

type placement struct {
	configBytes []byte

	config Config
	tag    string
}

const (
	TagAnnotation = "rigdev.placement/tag"
)

func (p *placement) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	p.tag = req.Tag
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
	if p.config.RequireTag {
		if p.tag == "" {
			return false
		}
		if v := capsule.Annotations[TagAnnotation]; v != p.tag {
			return false
		}
	}
	return true
}

func main() {
	plugin.StartPlugin("rigdev.placement", &placement{})
}
