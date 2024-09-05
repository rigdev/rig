package v1alpha2

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CapsuleSpec defines the desired state of Capsule
type CapsuleSpec struct {
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

	// Files is a list of files to mount in the container. These can either be
	// based on ConfigMaps or Secrets.
	Files []File `json:"files,omitempty"`

	// Scale specifies the scaling of the Capsule.
	Scale CapsuleScale `json:"scale,omitempty"`

	// NodeSelector is a selector for what nodes the Capsule should live on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Env specifies configuration for how the container should obtain
	// environment variables.
	Env Env `json:"env,omitempty"`

	CronJobs []CronJob `json:"cronJobs,omitempty"`
}

type CronJob struct {
	// +kubebuilder:validation:Required
	Name string `json:"name" protobuf:"1"`
	// +kubebuilder:validation:Required
	Schedule string `json:"schedule" protobuf:"2"`

	URL     *URL        `json:"url,omitempty" protobuf:"3"`
	Command *JobCommand `json:"command,omitempty" protobuf:"4"`
	// Defaults to 6
	MaxRetries     *uint `json:"maxRetries,omitempty" protobuf:"5"`
	TimeoutSeconds *uint `json:"timeoutSeconds,omitempty" protobuf:"6"`
}

type URL struct {
	// +kubebuilder:validation:Required
	Port uint16 `json:"port" protobuf:"1"`
	// +kubebuilder:validation:Required
	Path            string            `json:"path" protobuf:"2"`
	QueryParameters map[string]string `json:"queryParameters,omitempty" protobuf:"3"`
}

type JobCommand struct {
	// +kubebuilder:validation:Required
	Command string   `json:"command" protobuf:"1"`
	Args    []string `json:"args,omitempty" protobuf:"2"`
}

// Env defines what secrets and configmaps should be used for environment
// variables in the capsule.
type Env struct {
	// DisableAutomatic sets wether the capsule should disable automatically use
	// of existing secrets and configmaps which share the same name as the capsule
	// as environment variables.
	DisableAutomatic bool `json:"disable_automatic,omitempty" protobuf:"1"`

	// From holds a list of references to secrets and configmaps which should
	// be mounted as environment variables.
	From []EnvReference `json:"from,omitempty" protobuf:"2"`
}

// EnvSource holds a reference to either a ConfigMap or a Secret
type EnvReference struct {
	// Kind is the resource kind of the env reference, must be ConfigMap or Secret.
	Kind string `json:"kind"`
	// Name is the name of a ConfigMap or Secret in the same namespace as the Capsule.
	Name string `json:"name"`
}

// CapsuleScale specifies the horizontal and vertical scaling of the Capsule.
type CapsuleScale struct {
	// Horizontal specifies the horizontal scaling of the Capsule.
	Horizontal HorizontalScale `json:"horizontal,omitempty" protobuf:"1"`

	// Vertical specifies the vertical scaling of the Capsule.
	Vertical *VerticalScale `json:"vertical,omitempty" protobuf:"2"`
}

// HorizontalScale defines the policy for the number of replicas of
// the capsule It can both be configured with autoscaling and with a
// static number of replicas
type HorizontalScale struct {
	// Instances specifies minimum and maximum amount of Capsule
	// instances.
	Instances Instances `json:"instances" protobuf:"1"`

	// CPUTarget specifies that this Capsule should be scaled using CPU
	// utilization.
	CPUTarget *CPUTarget `json:"cpuTarget,omitempty" protobuf:"2"`
	// CustomMetrics specifies custom metrics emitted by the custom.metrics.k8s.io API
	// which the autoscaler should scale on
	CustomMetrics []CustomMetric `json:"customMetrics,omitempty" protobuf:"3" patchStrategy:"replace"`
}

// Instances specifies the minimum and maximum amount of capsule
// instances.
type Instances struct {
	// Min specifies the minimum amount of instances to run.
	Min uint32 `json:"min" protobuf:"1"`

	// Max specifies the maximum amount of instances to run. Omit to
	// disable autoscaling.
	Max *uint32 `json:"max,omitempty" protobuf:"2"`
}

// CPUTarget defines an autoscaler target for the CPU metric
// If empty, no autoscaling will be done
type CPUTarget struct {
	// Utilization specifies the average CPU target. If the average
	// exceeds this number new instances will be added.
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=100
	Utilization *uint32 `json:"utilization,omitempty" protobuf:"1"`
}

// CustomMetric defines a custom metrics emitted by the custom.metrics.k8s.io API
// which the autoscaler should scale on
// Exactly one of InstanceMetric and ObjectMetric must be provided
type CustomMetric struct {
	// InstanceMetric defines a custom instance-based metric (pod-metric in Kubernetes lingo)
	InstanceMetric *InstanceMetric `json:"instanceMetric,omitempty" protobuf:"1"`
	// ObjectMetric defines a custom object-based metric
	ObjectMetric *ObjectMetric `json:"objectMetric,omitempty" protobuf:"2"`
}

// InstanceMetric defines a custom instance-based metric (pod-metric in Kubernetes lingo)
type InstanceMetric struct {
	// +kubebuilder:validation:Required
	// MetricName is the name of the metric
	MetricName string `json:"metricName" protobuf:"1"`
	// MatchLabels is a set of key, value pairs which filters the metric series
	MatchLabels map[string]string `json:"matchLabels,omitempty" protobuf:"2"`
	// +kubebuilder:validation:Required
	// AverageValue defines the average value across all instances which the autoscaler scales towards
	AverageValue string `json:"averageValue" protobuf:"3"`
}

// ObjectMetric defines a custom object metric for the autoscaler
type ObjectMetric struct {
	// +kubebuilder:validation:Required
	// MetricName is the name of the metric
	MetricName string `json:"metricName" protobuf:"1"`
	// MatchLabels is a set of key, value pairs which filters the metric series
	MatchLabels map[string]string `json:"matchLabels,omitempty" protobuf:"2"`
	// AverageValue scales the number of instances towards making the value returned by the metric
	// divided by the number of instances reach AverageValue
	// Exactly one of 'Value' and 'AverageValue' must be set
	AverageValue string `json:"averageValue,omitempty" protobuf:"3"`
	// Value scales the number of instances towards making the value returned by the metric 'Value'
	// Exactly one of 'Value' and 'AverageValue' must be set
	Value string `json:"value,omitempty" protobuf:"4"`
	// +kubebuilder:validation:Required
	// DescribedObject is a reference to the object in the same namespace which is described by the metric
	DescribedObject autoscalingv2.CrossVersionObjectReference `json:"objectReference" protobuf:"5"`
}

// VerticalScale specifies the vertical scaling of the Capsule.
type VerticalScale struct {
	// CPU specifies the CPU resource request and limit
	CPU *ResourceLimits `json:"cpu,omitempty" protobuf:"1"`

	// Memory specifies the Memory resource request and limit
	Memory *ResourceLimits `json:"memory,omitempty" protobuf:"2"`

	// GPU specifies the GPU resource request and limit
	GPU *ResourceRequest `json:"gpu,omitempty" protobuf:"3"`
}

// ResourceLimits specifies the request and limit of a resource.
type ResourceLimits struct {
	// Request specifies the resource request.
	Request *resource.Quantity `json:"request,omitempty" protobuf:"1"`
	// Limit specifies the resource limit.
	Limit *resource.Quantity `json:"limit,omitempty" protobuf:"2"`
}

// ResourceRequest specifies the request of a resource.
type ResourceRequest struct {
	// Request specifies the request of a resource.
	Request resource.Quantity `json:"request,omitempty" protobuf:"1"`
}

// CapsuleInterface defines an interface for a capsule
type CapsuleInterface struct {
	// Name specifies a descriptive name of the interface.
	Name string `json:"name" protobuf:"1"`

	// Port specifies what port the interface should have.
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	Port int32 `json:"port" protobuf:"2"`

	// Liveness specifies that this interface should be used for
	// liveness probing. Only one of the Capsule interfaces can be
	// used as liveness probe.
	Liveness *InterfaceLivenessProbe `json:"liveness,omitempty" protobuf:"3"`

	// Readiness specifies that this interface should be used for
	// readiness probing. Only one of the Capsule interfaces can be
	// used as readiness probe.
	Readiness *InterfaceReadinessProbe `json:"readiness,omitempty" protobuf:"4"`

	// Public specifies if and how the interface should be published.
	Public *CapsulePublicInterface `json:"public,omitempty" protobuf:"5"`

	// Host routes that are mapped to this interface.
	Routes []HostRoute `json:"routes,omitempty" protobuf:"6"`
}

// HostRoute is the configuration of a route to the network interface
// it's configured on.
type HostRoute struct {
	// ID of the route. This field is required and cannot be empty, and must be unique for the interface.
	// If this field is changed, it may result in downtime, as it is used to generate resources.
	ID string `json:"id" protobuf:"1"`
	// Host of the route. This field is required and cannot be empty.
	Host string `json:"host" protobuf:"2"`
	// HTTP paths of the host that maps to the interface. If empty, all paths are
	// automatically matched.
	Paths []HTTPPathRoute `json:"paths,omitempty" protobuf:"3"`
	// Options for all paths of this host.
	RouteOptions `json:",inline"`
}

// PathMatchType specifies the semantics of how HTTP paths should be compared.
type PathMatchType string

const (
	// Exact match type, for when the path should match exactly.
	Exact PathMatchType = "Exact"
	// Path prefix, for when only the prefix needs to match.
	PathPrefix PathMatchType = "PathPrefix"
	// Path regular expression, for when the path should match a regular expression.
	RegularExpression PathMatchType = "RegularExpression"
)

// A HTTP path routing.
type HTTPPathRoute struct {
	// Path of the route.
	Path string `json:"path" protobuf:"1"`
	// The method of matching. By default, `PathPrefix` is used.
	// +kubebuilder:validation:Enum=PathPrefix;Exact;RegularExpression
	Match PathMatchType `json:"match,omitempty" protobuf:"2"`
}

// Route options.
type RouteOptions struct {
	// Annotations of the route option. This can be plugin-specific configuration
	// that allows custom plugins to add non-standard behavior.
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"4"`
}

// InterfaceLivenessProbe specifies an interface probe for liveness checks.
type InterfaceLivenessProbe struct {
	// Path is the HTTP path of the probe. Path is mutually
	// exclusive with the TCP and GCRP fields.
	Path string `json:"path,omitempty" protobuf:"1"`

	// TCP specifies that this is a simple TCP listen probe.
	TCP bool `json:"tcp,omitempty" protobuf:"2"`

	// GRPC specifies that this is a GRCP probe.
	GRPC *InterfaceGRPCProbe `json:"grpc,omitempty" protobuf:"3"`

	// For slow-starting containers, the startup delay allows liveness
	// checks to fail for a set duration before restarting the instance.
	StartupDelay uint32 `json:"startupDelay,omitempty" protobuf:"4"`
}

func (p InterfaceLivenessProbe) GetPath() string              { return p.Path }
func (p InterfaceLivenessProbe) GetTCP() bool                 { return p.TCP }
func (p InterfaceLivenessProbe) GetGRPC() *InterfaceGRPCProbe { return p.GRPC }

// InterfaceReadinessProbe specifies an interface probe for readiness checks.
type InterfaceReadinessProbe struct {
	// Path is the HTTP path of the probe. Path is mutually
	// exclusive with the TCP and GCRP fields.
	Path string `json:"path,omitempty" protobuf:"1"`

	// TCP specifies that this is a simple TCP listen probe.
	TCP bool `json:"tcp,omitempty" protobuf:"2"`

	// GRPC specifies that this is a GRCP probe.
	GRPC *InterfaceGRPCProbe `json:"grpc,omitempty" protobuf:"3"`
}

func (p InterfaceReadinessProbe) GetPath() string              { return p.Path }
func (p InterfaceReadinessProbe) GetTCP() bool                 { return p.TCP }
func (p InterfaceReadinessProbe) GetGRPC() *InterfaceGRPCProbe { return p.GRPC }

// InterfaceGRPCProbe specifies a GRPC probe.
type InterfaceGRPCProbe struct {
	// Service specifies the gRPC health probe service to probe. This is a
	// used as service name as per standard gRPC health/v1.
	Service string `json:"service" protobuf:"1"`

	// Enabled controls if the gRPC health check is activated.
	Enabled bool `json:"enabled,omitempty" protobuf:"2"`
}

// CapsulePublicInterface defines how to publicly expose the interface
type CapsulePublicInterface struct {
	// Ingress specifies that this interface should be exposed through an
	// Ingress resource. The Ingress field is mutually exclusive with the
	// LoadBalancer field.
	Ingress *CapsuleInterfaceIngress `json:"ingress,omitempty" protobuf:"1"`

	// LoadBalancer specifies that this interface should be exposed through a
	// LoadBalancer Service. The LoadBalancer field is mutually exclusive with
	// the Ingress field.
	LoadBalancer *CapsuleInterfaceLoadBalancer `json:"loadBalancer,omitempty" protobuf:"2"`
}

// CapsuleInterfaceIngress defines that the interface should be exposed as http
// ingress
type CapsuleInterfaceIngress struct {
	// Host specifies the DNS name of the Ingress resource.
	Host string `json:"host" protobuf:"1"`

	// Paths specifies a list of paths. In order for a request to
	// hit the ingress at least one of these must match the request.
	Paths []string `json:"paths,omitempty" protobuf:"2"`
}

// CapsuleInterfaceLoadBalancer defines that the interface should be exposed as
// a L4 loadbalancer
type CapsuleInterfaceLoadBalancer struct {
	// Port is the external port on the LoadBalancer
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	Port int32 `json:"port" protobuf:"1"`
}

// File defines a mounted file and where to retrieve the contents from
type File struct {
	// Ref specifies a reference to a ConfigMap or Secret key which holds the contents of the file.
	Ref *FileContentReference `json:"ref,omitempty"`

	// Path specifies the full path where the File should be mounted including
	// the file name.
	Path string `json:"path"`
}

// FileContentRef defines the name of a config resource and the key from which
// to retrieve the contents
type FileContentReference struct {
	// Kind of reference. Can be either ConfigMap or Secret.
	Kind string `json:"kind"`

	// Name of reference.
	Name string `json:"name"`

	// Key in reference which holds file contents.
	Key string `json:"key"`
}

// CapsuleStatus defines the observed state of Capsule
type CapsuleStatus struct {
	Replicas           uint32            `json:"replicas,omitempty"`
	ObservedGeneration int64             `json:"observedGeneration,omitempty"`
	OwnedResources     []OwnedResource   `json:"ownedResources,omitempty"`
	UsedResources      []UsedResource    `json:"usedResources,omitempty"`
	Deployment         *DeploymentStatus `json:"deploymentStatus,omitempty"`
	Errors             []string          `json:"errors,omitempty"`
}

type DeploymentStatus struct {
	// +kubebuilder:validation:Enum=created;failed
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type OwnedResource struct {
	Ref *v1.TypedLocalObjectReference `json:"ref"`
	// +kubebuilder:validation:Enum=created;failed;alreadyExists;unchanged;updated;changePending;deleted
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type UsedResource struct {
	Ref *v1.TypedLocalObjectReference `json:"ref"`
	// +kubebuilder:validation:Enum=found;missing;error
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

	// Spec holds the specification of the Capsule.
	Spec CapsuleSpec `json:"spec,omitempty"`

	// Status holds the status of the Capsule
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
