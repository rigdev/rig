package controller

import (
	"context"
	"fmt"
	"slices"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/ptr"
	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
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

	if capsuleHasIngress(req) {
		if !ingressIsSupported(s.cfg) {
			return errors.New("ingress is not supported")
		}

		ingresses := s.createIngresses(req)
		for _, ing := range ingresses {
			if err := req.Set(ing); err != nil {
				return err
			}
		}

		if shouldCreateCertificateResource(s.cfg) {
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

func (s *NetworkStep) createCertificate(req pipeline.CapsuleRequest) *cmv1.Certificate {
	crt := &cmv1.Certificate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Certificate",
			APIVersion: "cert-manager.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
		},
		Spec: cmv1.CertificateSpec{
			SecretName: fmt.Sprintf("%s-tls", req.Capsule().Name),
		},
	}

	if s.cfg.Certmanager != nil {
		crt.Spec.IssuerRef = cmmetav1.ObjectReference{
			Kind: cmv1.ClusterIssuerKind,
			Name: s.cfg.Certmanager.ClusterIssuer,
		}
	}

	for _, inf := range req.Capsule().Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
			crt.Spec.DNSNames = append(crt.Spec.DNSNames, inf.Public.Ingress.Host)
		}

		for _, route := range inf.Routes {
			if !slices.Contains(crt.Spec.DNSNames, route.Host) {
				crt.Spec.DNSNames = append(crt.Spec.DNSNames, route.Host)
			}
		}
	}

	return crt
}

func (s *NetworkStep) createIngresses(req pipeline.CapsuleRequest) []*netv1.Ingress {
	var ingresses []*netv1.Ingress

	// Public interface ingress
	var publicIngress *netv1.Ingress
	for _, inf := range req.Capsule().Spec.Interfaces {
		if inf.Public == nil || inf.Public.Ingress == nil {
			continue
		} else if publicIngress == nil {
			publicIngress = createBasicIngress(req, s.cfg, req.Capsule().Name)
		}

		rule := netv1.IngressRule{
			Host: inf.Public.Ingress.Host,
			IngressRuleValue: netv1.IngressRuleValue{
				HTTP: &netv1.HTTPIngressRuleValue{},
			},
		}

		if len(inf.Public.Ingress.Paths) == 0 {
			path := ""
			if s.cfg.Ingress.PathType == netv1.PathTypeExact || s.cfg.Ingress.PathType == netv1.PathTypePrefix {
				path = "/"
			}
			rule.IngressRuleValue.HTTP.Paths = []netv1.HTTPIngressPath{
				{
					PathType: ptr.New(s.cfg.Ingress.PathType),
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
				rule.IngressRuleValue.HTTP.Paths = append(
					rule.IngressRuleValue.HTTP.Paths,
					netv1.HTTPIngressPath{
						PathType: ptr.New(s.cfg.Ingress.PathType),
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

		if !s.cfg.Ingress.IsTLSDisabled() && inf.Public.Ingress.Host != "" {
			if len(publicIngress.Spec.TLS) == 0 {
				publicIngress.Spec.TLS = []netv1.IngressTLS{{
					SecretName: fmt.Sprintf("%s-tls", req.Capsule().Name),
				}}
			}
			publicIngress.Spec.TLS[0].Hosts = append(publicIngress.Spec.TLS[0].Hosts, inf.Public.Ingress.Host)
		}

		publicIngress.Spec.Rules = append(publicIngress.Spec.Rules, rule)
	}

	if publicIngress != nil {
		ingresses = append(ingresses, publicIngress)
	}

	// interface routes
	for _, inf := range req.Capsule().Spec.Interfaces {
		for _, route := range inf.Routes {
			ing := createBasicIngress(req, s.cfg, fmt.Sprintf("%s-%s", req.Capsule().Name, route.ID))
			rule := netv1.IngressRule{
				Host: route.Host,
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{},
				},
			}

			for key, value := range route.Annotations {
				ing.Annotations[key] = value
			}

			if len(route.Paths) == 0 {
				rule.IngressRuleValue.HTTP.Paths = []netv1.HTTPIngressPath{
					{
						PathType: ptr.New(netv1.PathTypePrefix),
						Path:     "/",
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
				for _, path := range route.Paths {
					match := netv1.PathTypePrefix
					if path.Match == v1alpha2.Exact {
						match = netv1.PathTypeExact
					}
					rule.IngressRuleValue.HTTP.Paths = append(
						rule.IngressRuleValue.HTTP.Paths,
						netv1.HTTPIngressPath{
							PathType: ptr.New(match),
							Path:     path.Path,
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

			if !s.cfg.Ingress.IsTLSDisabled() && route.Host != "" {
				if len(ing.Spec.TLS) == 0 {
					ing.Spec.TLS = []netv1.IngressTLS{{
						SecretName: fmt.Sprintf("%s-tls", req.Capsule().Name),
					}}
				}
				ing.Spec.TLS[0].Hosts = append(ing.Spec.TLS[0].Hosts, route.Host)
			}

			ing.Spec.Rules = append(ing.Spec.Rules, rule)
			ingresses = append(ingresses, ing)
		}
	}

	return ingresses
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

func shouldCreateCertificateResource(cfg *v1alpha1.OperatorConfig) bool {
	return cfg.Certmanager != nil &&
		cfg.Certmanager.CreateCertificateResources &&
		!cfg.Ingress.IsTLSDisabled()
}

func ingressIsSupported(cfg *v1alpha1.OperatorConfig) bool {
	return cfg.Ingress.IsTLSDisabled() ||
		(cfg.Certmanager != nil && cfg.Certmanager.ClusterIssuer != "")
}

func capsuleHasIngress(req pipeline.CapsuleRequest) bool {
	for _, inf := range req.Capsule().Spec.Interfaces {
		if (inf.Public != nil && inf.Public.Ingress != nil) || (len(inf.Routes) > 0) {
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

func createBasicIngress(req pipeline.CapsuleRequest, cfg *v1alpha1.OperatorConfig, name string) *netv1.Ingress {
	var ingressClassName *string
	if cfg.Ingress.ClassName != "" {
		ingressClassName = ptr.New(cfg.Ingress.ClassName)
	}

	i := &netv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   req.Capsule().Namespace,
			Annotations: maps.Clone(cfg.Ingress.Annotations),
		},
	}

	if i.Annotations == nil {
		i.Annotations = make(map[string]string)
	}

	if ingressClassName != nil {
		i.Spec.IngressClassName = ingressClassName
	}

	if ingressIsSupported(cfg) && !cfg.Ingress.IsTLSDisabled() && !shouldCreateCertificateResource(cfg) {
		i.Annotations["cert-manager.io/cluster-issuer"] = cfg.Certmanager.ClusterIssuer
	}

	return i
}
