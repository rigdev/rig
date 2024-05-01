// +groupName=plugins.rig.dev -- Only used for config doc generation
//
//nolint:revive
package service_monitor

import (
	"context"

	"github.com/hashicorp/go-hclog"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/pipeline"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "rigdev.service_monitor"
)

// Configuration for the deployment plugin
// +kubebuilder:object:root=true
type Config struct {
	Path     string
	PortName string
}

type Plugin struct {
	configBytes []byte
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error {
	// We do not have any configuration for this step?
	var cfg Config
	var err error
	if len(p.configBytes) > 0 {
		cfg, err = plugin.ParseTemplatedConfig[Config](p.configBytes, req, plugin.CapsuleStep[Config])
		if err != nil {
			return err
		}
	}

	// Consider returning an error. If you get this far, you should have a configuration.
	if cfg.PortName == "" {
		return errors.InvalidArgumentErrorf("portName is required to create a ServiceMonitor")
	}

	serviceMonitor := p.createPrometheusServiceMonitor(req, cfg)
	return req.Set(serviceMonitor)
}

func (p *Plugin) createPrometheusServiceMonitor(req pipeline.CapsuleRequest, cfg Config) *monitorv1.ServiceMonitor {
	return &monitorv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceMonitor",
			APIVersion: "monitoring.coreos.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            req.Capsule().Name,
			Namespace:       req.Capsule().Namespace,
			ResourceVersion: "",
			Labels: map[string]string{
				pipeline.LabelCapsule: req.Capsule().Name,
			},
		},
		Spec: monitorv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					pipeline.LabelCapsule: req.Capsule().Name,
				},
			},
			Endpoints: []monitorv1.Endpoint{{
				Port: cfg.PortName,
				Path: cfg.Path,
			}},
		},
	}
}
