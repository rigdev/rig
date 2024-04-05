// +groupName=plugins.rig.dev -- Only used for config doc generation
package ingressroutes

import (
	"context"
	"errors"
	"fmt"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/ptr"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const Name = "rigdev.ingress_routes"

// Configuration for the ingress_routes plugin
// +kubebuilder:object:root=true
type Config struct {
	// ClusterIssuer to use for issueing ingress certificates
	ClusterIssuer string `json:"clusterIssuer,omitempty"`

	// CreateCertificateResources specifies wether to create Certificate
	// resources. If this is not enabled we will use ingress annotations. This
	// is handy in environments where the ingress-shim isn't enabled.
	CreateCertificateResources bool `json:"createCertificateResources,omitempty"`

	// ClassName specifies the default ingress class to use for all ingress
	// resources created.
	IngressClassName string `json:"ingressClassName,omitempty"`

	// DisableTLS for ingress resources generated. This is useful if a 3rd-party component
	// is handling the HTTPS TLS termination and certificates.
	DisableTLS bool `json:"disableTLS,omitempty"`
}

type Plugin struct {
	configBytes []byte
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	var config Config
	var err error
	if len(p.configBytes) > 0 {
		config, err = plugin.ParseTemplatedConfig[Config](p.configBytes, req, plugin.CapsuleStep[Config])
		if err != nil {
			return err
		}
	}

	if capsuleHasIngress(req) {
		if !ingressIsSupported(config) {
			return errors.New("ingress is not supported. Either disable TLS or set a cluster issuer")
		}

		ingresses := p.createIngresses(req, config)
		for _, ing := range ingresses {
			if err := req.Set(ing); err != nil {
				return err
			}
		}

		if shouldCreateCertificateResource(config) {
			for _, crt := range p.createCertificate(req, config) {
				if err := req.Set(crt); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func shouldCreateCertificateResource(cfg Config) bool {
	return cfg.CreateCertificateResources && !cfg.DisableTLS
}

func ingressIsSupported(cfg Config) bool {
	return cfg.DisableTLS || cfg.ClusterIssuer != ""
}

func capsuleHasIngress(req pipeline.CapsuleRequest) bool {
	for _, inf := range req.Capsule().Spec.Interfaces {
		if (inf.Public != nil && inf.Public.Ingress != nil) || (len(inf.Routes) > 0) {
			return true
		}
	}
	return false
}

func getRoutes(inf v1alpha2.CapsuleInterface) []v1alpha2.HostRoute {
	routes := inf.Routes
	if inf.Public != nil && inf.Public.Ingress != nil {
		paths := []v1alpha2.HTTPPathRoute{}
		for _, path := range inf.Public.Ingress.Paths {
			paths = append(paths, v1alpha2.HTTPPathRoute{
				Path:  path,
				Match: v1alpha2.PathPrefix,
			})
		}

		routes = append(routes, v1alpha2.HostRoute{
			ID:    "public",
			Host:  inf.Public.Ingress.Host,
			Paths: paths,
		})
	}
	return routes
}

func getRouteName(req pipeline.CapsuleRequest, route v1alpha2.HostRoute) string {
	return fmt.Sprintf("%s-%s", req.Capsule().Name, route.ID)
}

func (p *Plugin) createCertificate(req pipeline.CapsuleRequest, cfg Config) []*cmv1.Certificate {
	var crts []*cmv1.Certificate

	for _, inf := range req.Capsule().Spec.Interfaces {
		for _, route := range getRoutes(inf) {
			name := getRouteName(req, route)

			crt := &cmv1.Certificate{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Certificate",
					APIVersion: "cert-manager.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: req.Capsule().Namespace,
				},
				Spec: cmv1.CertificateSpec{
					SecretName: fmt.Sprintf("%s-tls", name),
					IssuerRef: cmmetav1.ObjectReference{
						Kind: cmv1.ClusterIssuerKind,
						Name: cfg.ClusterIssuer,
					},
					DNSNames: []string{route.Host},
				},
			}

			crts = append(crts, crt)
		}
	}

	return crts
}

func (p *Plugin) createIngresses(req pipeline.CapsuleRequest, cfg Config) []*netv1.Ingress {
	var ingresses []*netv1.Ingress
	for _, inf := range req.Capsule().Spec.Interfaces {
		for _, route := range getRoutes(inf) {
			name := getRouteName(req, route)
			ing := createBasicIngress(req, cfg, name)
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
					var pt *netv1.PathType
					switch path.Match {
					case v1alpha2.Exact:
						pt = ptr.New(netv1.PathTypeExact)
					default:
						pt = ptr.New(netv1.PathTypePrefix)
					}

					rule.IngressRuleValue.HTTP.Paths = append(
						rule.IngressRuleValue.HTTP.Paths,
						netv1.HTTPIngressPath{
							PathType: pt,
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

			if !cfg.DisableTLS && route.Host != "" {
				if len(ing.Spec.TLS) == 0 {
					ing.Spec.TLS = []netv1.IngressTLS{{
						SecretName: fmt.Sprintf("%s-tls", name),
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

func createBasicIngress(req pipeline.CapsuleRequest, cfg Config, name string) *netv1.Ingress {
	var ingressClassName *string
	if cfg.IngressClassName != "" {
		ingressClassName = ptr.New(cfg.IngressClassName)
	}

	i := &netv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   req.Capsule().Namespace,
			Annotations: map[string]string{},
		},
	}

	if ingressClassName != nil {
		i.Spec.IngressClassName = ingressClassName
	}

	if ingressIsSupported(cfg) && !cfg.DisableTLS && !shouldCreateCertificateResource(cfg) {
		i.Annotations["cert-manager.io/cluster-issuer"] = cfg.ClusterIssuer
	}

	return i
}
