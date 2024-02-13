package controller

import (
	"context"
	"fmt"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type NetworkStep struct{}

func NewNetworkStep() *NetworkStep {
	return &NetworkStep{}
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

	if ingressIsSupported(req) && capsuleHasIngress(req) {
		if err := req.Set(s.createIngress(req)); err != nil {
			return err
		}
		if shouldCreateCertificateResource(req) {
			if err := req.Set(s.createCertificate(req)); err != nil {
				return err
			}
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
			Type: req.Config().Service.Type,
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

func (s *NetworkStep) createCertificate(req pipeline.CapsuleRequest) *cmv1.Certificate {
	crt := &cmv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
		},
		Spec: cmv1.CertificateSpec{
			SecretName: fmt.Sprintf("%s-tls", req.Capsule().Name),
		},
	}

	if req.Config().Certmanager != nil {
		crt.Spec.IssuerRef = cmmetav1.ObjectReference{
			Kind: cmv1.ClusterIssuerKind,
			Name: req.Config().Certmanager.ClusterIssuer,
		}
	}

	for _, inf := range req.Capsule().Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
			crt.Spec.DNSNames = append(crt.Spec.DNSNames, inf.Public.Ingress.Host)
		}
	}

	return crt
}

func (s *NetworkStep) createIngress(req pipeline.CapsuleRequest) *netv1.Ingress {
	ing := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Capsule().Name,
			Namespace:   req.Capsule().Namespace,
			Annotations: req.Config().Ingress.Annotations,
		},
	}

	if req.Config().Ingress.ClassName != "" {
		ing.Spec.IngressClassName = ptr.New(req.Config().Ingress.ClassName)
	}

	if ingressIsSupported(req) && !req.Config().Ingress.IsTLSDisabled() && !shouldCreateCertificateResource(req) {
		ing.Annotations["cert-manager.io/cluster-issuer"] = req.Config().Certmanager.ClusterIssuer
	}

	for _, inf := range req.Capsule().Spec.Interfaces {
		if inf.Public == nil || inf.Public.Ingress == nil {
			continue
		}

		ing.Spec.Rules = append(ing.Spec.Rules, netv1.IngressRule{
			Host: inf.Public.Ingress.Host,
			IngressRuleValue: netv1.IngressRuleValue{
				HTTP: &netv1.HTTPIngressRuleValue{},
			},
		})

		if len(inf.Public.Ingress.Paths) == 0 {
			path := ""
			if req.Config().Ingress.PathType == netv1.PathTypeExact || req.Config().Ingress.PathType == netv1.PathTypePrefix {
				path = "/"
			}
			ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths = []netv1.HTTPIngressPath{
				{
					PathType: ptr.New(req.Config().Ingress.PathType),
					Path:     path,
					Backend: netv1.IngressBackend{
						Service: &netv1.IngressServiceBackend{
							Name: req.Capsule().Name,
							Port: netv1.ServiceBackendPort{
								Name: inf.Name,
							},
						},
					},
				},
			}
		} else {
			for _, path := range inf.Public.Ingress.Paths {
				ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths = append(
					ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths,
					netv1.HTTPIngressPath{
						PathType: ptr.New(req.Config().Ingress.PathType),
						Path:     path,
						Backend: netv1.IngressBackend{
							Service: &netv1.IngressServiceBackend{
								Name: req.Capsule().Name,
								Port: netv1.ServiceBackendPort{
									Name: inf.Name,
								},
							},
						},
					},
				)
			}
		}

		if !req.Config().Ingress.IsTLSDisabled() && inf.Public.Ingress.Host != "" {
			if len(ing.Spec.TLS) == 0 {
				ing.Spec.TLS = []netv1.IngressTLS{{
					SecretName: fmt.Sprintf("%s-tls", req.Capsule().Name),
				}}
			}
			ing.Spec.TLS[0].Hosts = append(ing.Spec.TLS[0].Hosts, inf.Public.Ingress.Host)
		}
	}

	return ing
}

func (s *NetworkStep) createLoadBalancer(req pipeline.CapsuleRequest) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-lb", req.Capsule().Name),
			Namespace: req.Capsule().Namespace,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Selector: map[string]string{
				LabelCapsule: req.Capsule().Name,
			},
		},
	}

	for _, inf := range req.Capsule().Spec.Interfaces {
		if inf.Public != nil && inf.Public.LoadBalancer != nil {
			svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
				Name:       inf.Name,
				Port:       inf.Public.LoadBalancer.Port,
				TargetPort: intstr.FromString(inf.Name),
			})
		}
	}

	return svc
}

func shouldCreateCertificateResource(req pipeline.CapsuleRequest) bool {
	return req.Config().Certmanager != nil &&
		req.Config().Certmanager.CreateCertificateResources &&
		!req.Config().Ingress.IsTLSDisabled()
}

func ingressIsSupported(req pipeline.CapsuleRequest) bool {
	return req.Config().Ingress.IsTLSDisabled() ||
		(req.Config().Certmanager != nil && req.Config().Certmanager.ClusterIssuer != "")
}

func capsuleHasIngress(req pipeline.CapsuleRequest) bool {
	for _, inf := range req.Capsule().Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
			return true
		}
	}
	return false
}

func capsuleHasLoadBalancer(req pipeline.CapsuleRequest) bool {
	for _, inf := range req.Capsule().Spec.Interfaces {
		if inf.Public != nil && inf.Public.LoadBalancer != nil {
			return true
		}
	}
	return false
}
