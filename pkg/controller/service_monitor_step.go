package controller

import (
	"context"

	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/pipeline"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceMonitorStep struct {
	cfg *v1alpha1.OperatorConfig
}

func NewServiceMonitorStep(cfg *v1alpha1.OperatorConfig) *ServiceMonitorStep {
	return &ServiceMonitorStep{
		cfg: cfg,
	}
}

func (s *ServiceMonitorStep) Apply(_ context.Context, req pipeline.CapsuleRequest) error {
	if s.cfg.PrometheusServiceMonitor == nil || s.cfg.PrometheusServiceMonitor.PortName == "" {
		return nil
	}

	serviceMonitor := s.createPrometheusServiceMonitor(req)
	return req.Set(serviceMonitor)
}

func (s *ServiceMonitorStep) createPrometheusServiceMonitor(req pipeline.CapsuleRequest) *monitorv1.ServiceMonitor {
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
				LabelCapsule: req.Capsule().Name,
			},
		},
		Spec: monitorv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					LabelCapsule: req.Capsule().Name,
				},
			},
			Endpoints: []monitorv1.Endpoint{{
				Port: s.cfg.PrometheusServiceMonitor.PortName,
				Path: s.cfg.PrometheusServiceMonitor.Path,
			}},
		},
	}
}
