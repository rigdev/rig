// +groupName=plugins.rig.dev -- Only used for config doc generation
//
//nolint:revive
package service_account

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/pipeline"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "rigdev.service_account"
)

// Configuration for the deployment plugin
// +kubebuilder:object:root=true
type Config struct{}

type Plugin struct {
	plugin.NoWatchObjectStatus

	configBytes []byte
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error {
	// We do not have any configuration for this step?
	// var config Config
	var err error
	if len(p.configBytes) > 0 {
		_, err = plugin.ParseTemplatedConfig[Config](p.configBytes, req, plugin.CapsuleStep[Config])
		if err != nil {
			return err
		}
	}

	sa := p.createServiceAccount(req)
	return req.Set(sa)
}

func (s *Plugin) createServiceAccount(req pipeline.CapsuleRequest) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
		},
	}

	return sa
}
