package controller

import (
	"context"
	"fmt"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type NetworkStep struct {
	cfg *v1alpha1.OperatorConfig
}

func NewNetworkStep(cfg *v1alpha1.OperatorConfig) *NetworkStep {
	return &NetworkStep{
		cfg: cfg,
	}
}

func (s *NetworkStep) Apply(_ context.Context, req pipeline.CapsuleRequest) error {
	// If no interfaces are defined, no changes are needed.
	if len(req.Capsule().Spec.Interfaces) == 0 {
		return nil
	}

	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); errors.IsNotFound(err) {
		// We assume service and ingress are not needed if the deployment doesn't exist.
		return nil
	} else if err != nil {
		return err
	}

	for i, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name != req.Capsule().Name {
			continue
		}

		var ports []corev1.ContainerPort
		for _, ni := range req.Capsule().Spec.Interfaces {
			ports = append(ports, corev1.ContainerPort{
				Name:          ni.Name,
				ContainerPort: ni.Port,
			})

			if ni.Liveness != nil {
				container.LivenessProbe = &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: ni.Liveness.Path,
							Port: intstr.FromInt32(ni.Port),
						},
					},
				}
			}

			if ni.Readiness != nil {
				container.ReadinessProbe = &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: ni.Readiness.Path,
							Port: intstr.FromInt32(ni.Port),
						},
					},
				}
			}
		}
		container.Ports = ports
		deployment.Spec.Template.Spec.Containers[i] = container
	}

	if err := req.Set(deployment); err != nil {
		return err
	}

	if err := req.Set(s.createService(req)); err != nil {
		return err
	}

	if capsuleHasLoadBalancer(req) {
		lb := s.createLoadBalancer(req)
		if err := req.Set(lb); err != nil {
			return err
		}
	}

	return nil
}

func (s *NetworkStep) createService(req pipeline.CapsuleRequest) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
			Labels: map[string]string{
				LabelCapsule: req.Capsule().Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				LabelCapsule: req.Capsule().Name,
			},
			Type: s.cfg.Service.Type,
		},
	}

	for _, inf := range req.Capsule().Spec.Interfaces {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       inf.Name,
			Port:       inf.Port,
			TargetPort: intstr.FromString(inf.Name),
		})
	}

	return svc
}

func (s *NetworkStep) createLoadBalancer(req pipeline.CapsuleRequest) *corev1.Service {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-lb", req.Capsule().Name),
			Namespace: req.Capsule().Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Selector: map[string]string{
				LabelCapsule: req.Capsule().Name,
			},
		},
	}

	for _, inf := range req.Capsule().Spec.Interfaces {
		if inf.Public != nil && inf.Public.LoadBalancer != nil {
			svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
				Name:       inf.Name,
				Port:       inf.Public.LoadBalancer.Port,
				TargetPort: intstr.FromString(inf.Name),
			})
		}
	}

	return svc
}

func capsuleHasLoadBalancer(req pipeline.CapsuleRequest) bool {
	for _, inf := range req.Capsule().Spec.Interfaces {
		if inf.Public != nil && inf.Public.LoadBalancer != nil {
			return true
		}
	}
	return false
}
