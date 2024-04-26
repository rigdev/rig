package pipeline

import (
	"context"

	"github.com/rigdev/rig/pkg/pipeline"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAccountStep struct{}

func NewServiceAccountStep() *ServiceAccountStep {
	return &ServiceAccountStep{}
}

func (s *ServiceAccountStep) Apply(_ context.Context, req pipeline.CapsuleRequest) error {
	sa := s.createServiceAccount(req)
	return req.Set(sa)
}

func (s *ServiceAccountStep) createServiceAccount(req pipeline.CapsuleRequest) *corev1.ServiceAccount {
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
