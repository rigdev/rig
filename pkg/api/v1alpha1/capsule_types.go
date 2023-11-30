package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CapsuleSpec defines the desired state of Capsule
type CapsuleSpec struct {
	// Replicas specifies how many replicas the Capsule should have.
	Replicas *int32 `json:"replicas,omitempty"`

	// Image specifies what image the Capsule should run.
	Image string `json:"image"`

	// Command is run as a command in the shell. If left unspecified, the
	// container will run using what is specified as ENTRYPOINT in the
	// Dockerfile.
	Command string `json:"command,omitempty"`

	// Args is a list of arguments either passed to the Command or if Command
	// is left empty the arguments will be passed to the ENTRYPOINT of the
	// docker image.
	Args []string `json:"args,omitempty"`

	// Interfaces specifies the list of interfaces the the container should
	// have. Specifying interfaces will create the corresponding kubernetes
	// Services and Ingresses depending on how the interface is configured.
	Interfaces []CapsuleInterface `json:"interfaces,omitempty"`

	// Env specifies configuration for how the container should obtain
	// environment variables.
	Env *Env `json:"env,omitempty"`

	// Files is a list of files to mount in the container. These can either be
	// based on ConfigMaps or Secrets.
	Files []File `json:"files,omitempty"`

	// Resources describes what resources the Capsule should have access to.
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// ImagePullSecret is a reference to a secret holding docker credentials
	// for the registry of the image.
	ImagePullSecret *v1.LocalObjectReference `json:"imagePullSecret,omitempty"`

	// HorizontalScale describes how the Capsule should scale out
	HorizontalScale HorizontalScale `json:"horizontalScale,omitempty"`

	// ServiceAccountName specifies the name of an existing ServiceAccount
	// which the Capsule should run as.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// NodeSelector is a selector for what nodes the Capsule should live on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// CapsuleInterface defines an interface for a capsule
type CapsuleInterface struct {
	// Name specifies a descriptive name of the interface.
	Name string `json:"name"`

	// Port specifies what port the interface should have.
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// Public specifies if and how the interface should be published.
	Public *CapsulePublicInterface `json:"public,omitempty"`
}

// CapsulePublicInterface defines how to publicly expose the interface
type CapsulePublicInterface struct {
	// Ingress specifies that this interface should be exposed through an
	// Ingress resource. The Ingress field is mutually exclusive with the
	// LoadBalancer field.
	Ingress *CapsuleInterfaceIngress `json:"ingress,omitempty"`

	// LoadBalancer specifies that this interface should be exposed through a
	// LoadBalancer Service. The LoadBalancer field is mutually exclusive with
	// the Ingress field.
	LoadBalancer *CapsuleInterfaceLoadBalancer `json:"loadBalancer,omitempty"`
}

// CapsuleInterfaceIngress defines that the interface should be exposed as http
// ingress
type CapsuleInterfaceIngress struct {
	// Host specifies the DNS name of the Ingress resource
	Host string `json:"host"`
}

// CapsuleInterfaceLoadBalancer defines that the interface should be exposed as
// a L4 loadbalancer
type CapsuleInterfaceLoadBalancer struct {
	// Port is the external port on the LoadBalancer
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// NodePort specifies a NodePort that the Service will use instead of
	// acting as a LoadBalancer.
	NodePort int32 `json:"nodePort,omitempty"`
}

// File defines a mounted file and where to retrieve the contents from
type File struct {
	// Path specifies the full path where the File should be mounted including
	// the file name.
	Path string `json:"path"`

	// ConfigMap specifies that this file is based on a key in a ConfigMap. The
	// ConfigMap field is mutually exclusive with Secret.
	ConfigMap *FileContentRef `json:"configMap,omitempty"`

	// Secret specifies that this file is based on a key in a Secret. The
	// Secret field is mutually exclusive with ConfigMap.
	Secret *FileContentRef `json:"secret,omitempty"`
}

// FileContentRef defines the name of a config resource and the key from which
// to retrieve the contents
type FileContentRef struct {
	// Name specifies the name of the Secret or ConfigMap.
	Name string `json:"name"`

	// Key specifies the key holding the file contents.
	Key string `json:"key"`
}

// Env defines what secrets and configmaps should be used for environment
// variables in the capsule.
type Env struct {
	// Automatic sets wether the capsule should automatically use existing
	// secrets and configmaps which share the same name as the capsule as
	// environment variables.
	Automatic *bool `json:"automatic,omitempty"`

	// From holds a list of references to secrets and configmaps which should
	// be mounted as environment variables.
	From []EnvSource `json:"from,omitempty"`
}

// EnvSource holds a reference to either a ConfigMap or a Secret
type EnvSource struct {
	// ConfigMapName is the name of a ConfigMap in the same namespace as the Capsule
	ConfigMapName string `json:"configMapName,omitempty"`

	// SecretName is the name of a Secret in the same namespace as the Capsule
	SecretName string `json:"secretName,omitempty"`
}

// HorizontalScale defines the policy for the number of replicas of the capsule
// It can both be configured with autoscaling and with a static number of replicas
type HorizontalScale struct {
	// MinReplicas is the minimum amount of replicas that the Capsule should
	// have.
	MinReplicas *uint32 `json:"minReplicas,omitempty"`

	// MaxReplicas is the maximum amount of replicas that the Capsule should
	// have.
	MaxReplicas *uint32 `json:"maxReplicas,omitempty"`

	// CPUTarget specifies that this Capsule should be scaled using CPU
	// utilization.
	CPUTarget CPUTarget `json:"cpuTarget,omitempty"`
}

// CPUTarget defines an autoscaler target for the CPU metric
// If empty, no autoscaling will be done
type CPUTarget struct {
	// AverageUtilizationPercentage sets the utilization which when exceeded
	// will trigger autoscaling.
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

	// Spec holds the specification of the Capsule.
	Spec CapsuleSpec `json:"spec,omitempty"`

	// Status holds the status of the Capsule
	Status CapsuleStatus `json:"status,omitempty"`

	// Scale holds metadata for the HorizontalPodAutoscaler
	Scale Scale `json:"scale,omitempty"`
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
