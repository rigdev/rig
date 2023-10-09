package v1alpha1

import (
	"github.com/rigdev/rig/pkg/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperatorConfig is the Schema for the configs API
// +kubebuilder:object:root=true
type OperatorConfig struct {
	metav1.TypeMeta `json:",inline"`

	// WebhooksEnabled set wether or not webhooks should be enabled. When
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

	// Ingress holds the configuration for ingress resources created by the
	// operator.
	Ingress IngressConfig `json:"ingress,omitempty"`
}

type CertManagerConfig struct {
	// ClusterIssuer to use for issueing ingress certificates
	ClusterIssuer string `json:"clusterIssuer"`

	// CreateCertificateResources specifies wether to create Certificate
	// resources. If this is not enabled we will use ingress annotations. This
	// is handy in environments where the ingress-shim isen't enabled.
	CreateCertificateResources bool `json:"createCertificateResources,omitempty"`
}

type IngressConfig struct {
	// Annotations for all ingress resources created.
	Annotations map[string]string `json:"annotations"`

	// ClassName specifies the default ingress class to use for all ingress
	// resources created.
	ClassName string `json:"className"`
}

func (c *OperatorConfig) Default() {
	if c.WebhooksEnabled == nil {
		c.WebhooksEnabled = ptr.New(true)
	}
	if c.LeaderElectionEnabled == nil {
		c.LeaderElectionEnabled = ptr.New(true)
	}
	if c.Ingress.Annotations == nil {
		c.Ingress.Annotations = map[string]string{}
	}
}

func init() {
	SchemeBuilder.Register(&OperatorConfig{})
}
