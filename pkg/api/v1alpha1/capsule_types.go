package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CapsuleSpec defines the desired state of Capsule
type CapsuleSpec struct {
	Replicas   int32                    `json:"replicas"`
	Image      string                   `json:"image"`
	Command    string                   `json:"command,omitempty"`
	Args       []string                 `json:"args,omitempty"`
	Interfaces []CapsuleInterface       `json:"interfaces,omitempty"`
	Files      []File                   `json:"files,omitempty"`
	Resources  *v1.ResourceRequirements `json:"resources,omitempty"`
}

// CapsuleInterface defines an interface for a capsule
type CapsuleInterface struct {
	Port   int32                   `json:"port"`
	Name   string                  `json:"name"`
	Public *CapsulePublicInterface `json:"public,omitempty"`
}

// CapsulePublicInterface defines how to publicly expose the interface
type CapsulePublicInterface struct {
	Ingress      *CapsuleInterfaceIngress      `json:"ingress,omitempty"`
	LoadBalancer *CapsuleInterfaceLoadBalancer `json:"loadBalancer,omitempty"`
}

// CapsuleInterfaceIngress defines that the interface should be exposed as http
// ingress
type CapsuleInterfaceIngress struct {
	Host string `json:"host"`
}

// CapsuleInterfaceLoadBalancer defines that the interface should be exposed as
// a L4 loadbalancer
type CapsuleInterfaceLoadBalancer struct {
	Port int32 `json:"port"`
}

// File defines a mounted file and where to retrieve the contents from
type File struct {
	Path      string          `json:"path"`
	ConfigMap *FileContentRef `json:"configMap,omitempty"`
	Secret    *FileContentRef `json:"secret,omitempty"`
}

// FileContentRef defines the name of a config resource and the key which from
// which to retrieve the contents
type FileContentRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// CapsuleStatus defines the observed state of Capsule
type CapsuleStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Capsule is the Schema for the capsules API
type Capsule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CapsuleSpec   `json:"spec,omitempty"`
	Status CapsuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CapsuleList contains a list of Capsule
type CapsuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Capsule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Capsule{}, &CapsuleList{})
}
