package controller

import (
	"context"

	"github.com/rigdev/rig/pkg/controller/pipeline"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAccountStep struct{}

func NewServiceAccountStep() *ServiceAccountStep {
	return &ServiceAccountStep{}
}

func (s *ServiceAccountStep) Apply(_ context.Context, req pipeline.Request) error {
	sa := s.createServiceAccount(req)
	req.Set(req.ObjectKey(pipeline.CoreServiceAccount), sa)
	return nil
}

func (s *ServiceAccountStep) createServiceAccount(req pipeline.Request) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
		},
	}

	return sa
}
