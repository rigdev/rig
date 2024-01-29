package controller

import (
	"context"

	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceMonitorStep struct{}

func NewServiceMonitorStep() *ServiceMonitorStep {
	return &ServiceMonitorStep{}
}

func (s *ServiceMonitorStep) Apply(ctx context.Context, req Request) error {
	if req.Config().PrometheusServiceMonitor == nil || req.Config().PrometheusServiceMonitor.PortName == "" {
		return nil
	}

	serviceMonitor := s.createPrometheusServiceMonitor(req)
	req.Set(req.ObjectKey(_monitoringServiceMonitorGVK), serviceMonitor)

	return nil
}

func (s *ServiceMonitorStep) createPrometheusServiceMonitor(req Request) *monitorv1.ServiceMonitor {
	return &monitorv1.ServiceMonitor{
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
