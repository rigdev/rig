package v1alpha1

import (
	"github.com/rigdev/rig/pkg/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperatorConfig is the Schema for the configs API
// +kubebuilder:object:root=true
type OperatorConfig struct {
	metav1.TypeMeta `json:",inline"`

	WebhooksEnabled       *bool `json:"webhooksEnabled,omitempty"`
	DevModeEnabled        bool  `json:"devModeEnabled,omitempty"`
	LeaderElectionEnabled *bool `json:"leaderElectionEnabled,omitempty"`
}

func (c *OperatorConfig) Default() {
	if c.WebhooksEnabled == nil {
		c.WebhooksEnabled = ptr.New(true)
	}
	if c.LeaderElectionEnabled == nil {
		c.LeaderElectionEnabled = ptr.New(true)
	}
}

func init() {
	SchemeBuilder.Register(&OperatorConfig{})
}
