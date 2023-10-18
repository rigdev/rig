package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CapsuleSpec defines the desired state of Capsule
type CapsuleSpec struct {
	Replicas *int32 `json:"replicas,omitempty"`
	Image    string `json:"image"`

	Command            string                   `json:"command,omitempty"`
	Args               []string                 `json:"args,omitempty"`
	Interfaces         []CapsuleInterface       `json:"interfaces,omitempty"`
	Files              []File                   `json:"files,omitempty"`
	Resources          *v1.ResourceRequirements `json:"resources,omitempty"`
	ImagePullSecret    *v1.LocalObjectReference `json:"imagePullSecret,omitempty"`
	HorizontalScale    HorizontalScale          `json:"horizontalScale,omitempty"`
	ServiceAccountName string                   `json:"serviceAccountName,omitempty"`
}

// CapsuleInterface defines an interface for a capsule
type CapsuleInterface struct {
	Name string `json:"name"`
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

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
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	Port     int32 `json:"port"`
	NodePort int32 `json:"nodePort,omitempty"`
}

// File defines a mounted file and where to retrieve the contents from
type File struct {
	Path      string          `json:"path"`
	ConfigMap *FileContentRef `json:"configMap,omitempty"`
	Secret    *FileContentRef `json:"secret,omitempty"`
}

// FileContentRef defines the name of a config resource and the key from which
// to retrieve the contents
type FileContentRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// HorizontalScale defines the policy for the number of replicas of the capsule
// It can both be configured with autoscaling and with a static number of replicas
type HorizontalScale struct {
	MinReplicas *uint32   `json:"minReplicas,omitempty"`
	MaxReplicas *uint32   `json:"maxReplicas,omitempty"`
	CPUTarget   CPUTarget `json:"cpuTarget,omitempty"`
}

// CPUTarget defines an autoscaler target for the CPU metric
// If empty, no autoscaling will be done
type CPUTarget struct {
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=100
	AverageUtilizationPercentage uint32 `json:"averageUtilizationPercentage"`
}

// CapsuleStatus defines the observed state of Capsule
type CapsuleStatus struct {
	Replicas           uint32           `json:"replicas,omitempty"`
	ObservedGeneration int64            `json:"observedGeneration,omitempty"`
	OwnedResources     []OwnedResource  `json:"ownedResources,omitempty"`
	Deployment         DeploymentStatus `json:"deploymentStatus,omitempty"`
}

type DeploymentStatus struct {
	// +kubebuilder:validation:Enum=created;failed
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type OwnedResource struct {
	Ref *v1.TypedLocalObjectReference `json:"ref"`
	// +kubebuilder:validation:Enum=created;failed
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type Scale struct {
	Replicas uint32 `json:"replicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas

// Capsule is the Schema for the capsules API
type Capsule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CapsuleSpec   `json:"spec,omitempty"`
	Status CapsuleStatus `json:"status,omitempty"`
	Scale  Scale         `json:"scale,omitempty"`
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
