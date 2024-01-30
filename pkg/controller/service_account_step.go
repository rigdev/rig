package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAccountStep struct{}

func NewServiceAccountStep() *ServiceAccountStep {
	return &ServiceAccountStep{}
}

func (s *ServiceAccountStep) Apply(_ context.Context, req Request) error {
	sa := s.createServiceAccount(req)
	req.Set(req.ObjectKey(_coreServiceAccount), sa)
	return nil
}

func (s *ServiceAccountStep) createServiceAccount(req Request) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
		},
	}

	return sa
}
