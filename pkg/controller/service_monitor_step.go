package controller

import (
	"context"

	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceMonitorStep struct{}

func NewServiceMonitorStep() *ServiceMonitorStep {
	return &ServiceMonitorStep{}
}

func (s *ServiceMonitorStep) Apply(_ context.Context, req pipeline.CapsuleRequest) error {
	if req.Config().PrometheusServiceMonitor == nil || req.Config().PrometheusServiceMonitor.PortName == "" {
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
				Port: req.Config().PrometheusServiceMonitor.PortName,
				Path: req.Config().PrometheusServiceMonitor.Path,
			}},
		},
	}
}
