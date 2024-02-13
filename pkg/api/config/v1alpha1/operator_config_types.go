package v1alpha1

import (
	"github.com/rigdev/rig/pkg/ptr"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	SchemeBuilder.Register(&OperatorConfig{})
}

// OperatorConfig is the Schema for the operator config API
// +kubebuilder:object:root=true
type OperatorConfig struct {
	metav1.TypeMeta `json:",inline"`

	// WebhooksEnabled sets wether or not webhooks should be enabled. When
	// enabled a certificate should be mounted at the webhook server
	// certificate path. Defaults to true if omitted.
	WebhooksEnabled *bool `json:"webhooksEnabled,omitempty"`

	// DevModeEnabled enables verbose logs and changes the logging format to be
	// more human readable.
	DevModeEnabled bool `json:"devModeEnabled,omitempty"`

	// LeaderElectionEnabled enables leader election when running multiple
	// instances of the operator.
	LeaderElectionEnabled *bool `json:"leaderElectionEnabled,omitempty"`

	// Certmanager holds configuration for how the operator should create
	// certificates for ingress resources.
	Certmanager *CertManagerConfig `json:"certManager,omitempty"`

	// Service holds the configuration for service resources created by the
	// operator.
	Service ServiceConfig `json:"service,omitempty"`

	// Ingress holds the configuration for ingress resources created by the
	// operator.
	Ingress IngressConfig `json:"ingress,omitempty"`

	// PrometheusServiceMonitor defines if Rig should spawn a Prometheus ServiceMonitor per capsule
	// for use with a Prometheus Operator stack.
	PrometheusServiceMonitor *PrometheusServiceMonitor `json:"prometheusServiceMonitor,omitempty"`

	// VerticalPodAutoscaler holds the configuration for the VerticalPodAutoscaler resources
	// potentially generated by the operator.
	VerticalPodAutoscaler VerticalPodAutoscaler `json:"verticalPodAutoscaler,omitempty"`

	// Steps to perform as part of running the operator.
	// +patchStrategy=merge
	Steps []Step `json:"steps,omitempty"`
}

type Step struct {
	// Plugin to use in the current step.
	Plugin string `json:"plugin,omitempty"`
	// If set, only capsules in one of the namespaces given will have this step run.
	Namespaces []string `json:"namespaces,omitempty"`
	Config     string   `json:"config,omitempty"`
}

type VerticalPodAutoscaler struct {
	// Enabled enables the creation of a VerticalPodAutoscaler per capsule
	Enabled bool `json:"enabled,omitempty"`
}

type PrometheusServiceMonitor struct {
	// Path is the path which Prometheus should query on ports. Defaults to /metrics if not set.
	Path string `json:"path,omitempty"`
	// PortName is the name of the port which Prometheus will query metrics on
	PortName string `json:"portName"`
}

type CertManagerConfig struct {
	// ClusterIssuer to use for issueing ingress certificates
	ClusterIssuer string `json:"clusterIssuer,omitempty"`

	// CreateCertificateResources specifies wether to create Certificate
	// resources. If this is not enabled we will use ingress annotations. This
	// is handy in environments where the ingress-shim isn't enabled.
	CreateCertificateResources bool `json:"createCertificateResources,omitempty"`
}

type ServiceConfig struct {
	// Type of the service to generate. By default, services are of type ClusterIP.
	// Valid values are ClusterIP, NodePort.
	Type corev1.ServiceType `json:"type,omitempty"`
}

type IngressConfig struct {
	// Annotations for all ingress resources created.
	Annotations map[string]string `json:"annotations,omitempty"`

	// ClassName specifies the default ingress class to use for all ingress
	// resources created.
	ClassName string `json:"className,omitempty"`

	// PathType defines how ingress paths should be interpreted.
	// Allowed values: Exact, Prefix, ImplementationSpecific
	PathType networkingv1.PathType `json:"pathType,omitempty"`

	// DisableTLS for ingress resources generated. This is useful if a 3rd-party component
	// is handling the HTTPS TLS termination and certificates.
	DisableTLS *bool `json:"disableTLS,omitempty"`
}

func (cfg IngressConfig) IsTLSDisabled() bool {
	return cfg.DisableTLS != nil && *cfg.DisableTLS
}

func (c *OperatorConfig) Default() *OperatorConfig {
	if c == nil {
		return c
	}
	c.SetGroupVersionKind(schema.FromAPIVersionAndKind(
		GroupVersion.Identifier(),
		"OperatorConfig",
	))
	if c.WebhooksEnabled == nil {
		c.WebhooksEnabled = ptr.New(true)
	}
	if c.LeaderElectionEnabled == nil {
		c.LeaderElectionEnabled = ptr.New(true)
	}
	if c.Service.Type == "" {
		c.Service.Type = corev1.ServiceTypeClusterIP
	}
	if c.Ingress.Annotations == nil {
		c.Ingress.Annotations = map[string]string{}
	}
	if c.Ingress.PathType == "" {
		c.Ingress.PathType = networkingv1.PathTypePrefix
	}
	return c
}
