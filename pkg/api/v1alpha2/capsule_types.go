package v1alpha2

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CapsuleSpec defines the desired state of Capsule
type CapsuleSpec struct {
	Image   string   `json:"image"`
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`

	Interfaces []CapsuleInterface `json:"interfaces,omitempty"`

	Files        []File            `json:"files,omitempty"`
	Scale        CapsuleScale      `json:"scale,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Env          *Env              `json:"env,omitempty"`
}

// Env defines what secrets and configmaps should be used for environment
// variables in the capsule.
type Env struct {
	// DisableAutomatic sets wether the capsule should disable automatically use
	// of existing secrets and configmaps which share the same name as the capsule
	// as environment variables.
	DisableAutomatic bool `json:"disable_automatic,omitempty"`

	// From holds a list of references to secrets and configmaps which should
	// be mounted as environment variables.
	From []EnvReference `json:"from,omitempty"`
}

// EnvSource holds a reference to either a ConfigMap or a Secret
type EnvReference struct {
	// Kind is the resource kind of the env reference, must be ConfigMap or Secret.
	Kind string `json:"kind"`
	// Name is the name of a ConfigMap or Secret in the same namespace as the Capsule.
	Name string `json:"name"`
}

type CapsuleScale struct {
	Horizontal HorizontalScale `json:"horizontal,omitempty"`
	Vertical   *VerticalScale  `json:"vertical,omitempty"`
}

// HorizontalScale defines the policy for the number of replicas of the capsule
// It can both be configured with autoscaling and with a static number of replicas
type HorizontalScale struct {
	Instances Instances  `json:"instances"`
	CPUTarget *CPUTarget `json:"cpuTarget,omitempty"`
}

type Instances struct {
	Min uint32  `json:"min"`
	Max *uint32 `json:"max,omitempty"`
}

// CPUTarget defines an autoscaler target for the CPU metric
// If empty, no autoscaling will be done
type CPUTarget struct {
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=100
	Utilization *uint32 `json:"utilization,omitempty"`
}

type VerticalScale struct {
	CPU    *ResourceLimits  `json:"cpu,omitempty"`
	Memory *ResourceLimits  `json:"memory,omitempty"`
	GPU    *ResourceRequest `json:"gpu,omitempty"`
}

type ResourceLimits struct {
	Request *resource.Quantity `json:"request,omitempty"`
	Limit   *resource.Quantity `json:"limit,omitempty"`
}

type ResourceRequest struct {
	Request resource.Quantity `json:"request,omitempty"`
}

// CapsuleInterface defines an interface for a capsule
type CapsuleInterface struct {
	Name string `json:"name"`
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	Liveness  *InterfaceProbe `json:"liveness,omitempty"`
	Readiness *InterfaceProbe `json:"readiness,omitempty"`

	Public *CapsulePublicInterface `json:"public,omitempty"`
}

type InterfaceProbe struct {
	Path string              `json:"path,omitempty"`
	TCP  bool                `json:"tcp,omitempty"`
	GRPC *InterfaceGRPCProbe `json:"grpc,omitempty"`
}

type InterfaceGRPCProbe struct {
	Service string `json:"service"`
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
	Port int32 `json:"port"`
}

// File defines a mounted file and where to retrieve the contents from
type File struct {
	Ref  *FileContentReference `json:"ref,omitempty"`
	Path string                `json:"path"`
}

// FileContentRef defines the name of a config resource and the key from which
// to retrieve the contents
type FileContentReference struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

// CapsuleStatus defines the observed state of Capsule
type CapsuleStatus struct {
	Replicas           uint32            `json:"replicas,omitempty"`
	ObservedGeneration int64             `json:"observedGeneration,omitempty"`
	OwnedResources     []OwnedResource   `json:"ownedResources,omitempty"`
	Deployment         *DeploymentStatus `json:"deploymentStatus,omitempty"`
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// Capsule is the Schema for the capsules API
type Capsule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CapsuleSpec    `json:"spec,omitempty"`
	Status *CapsuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:storageversion

// CapsuleList contains a list of Capsule
type CapsuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Capsule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Capsule{}, &CapsuleList{})
}
