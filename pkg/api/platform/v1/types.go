// +kubebuilder:object:generate=true
// +groupName=rig.platform
package v1

import (
	"maps"
	"slices"

	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/ptr"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type Environment struct {
	metav1.TypeMeta `json:",inline"`
	// Name is unique
	Name              string `json:"name" protobuf:"3"`
	NamespaceTemplate string `json:"namespaceTemplate" protobuf:"4"`
	OperatorVersion   string `json:"operatorVersion" protobuf:"5"`
	Cluster           string `json:"cluster" protobuf:"6"`
	// Environment level defaults
	Spec           ProjEnvCapsuleBase `json:"spec" protobuf:"7"`
	Ephemeral      bool               `json:"ephemeral" protobuf:"8"`
	ActiveProjects []string           `json:"activeProjects" protobuf:"9"`
	Global         bool               `json:"global" protobuf:"10"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type Project struct {
	metav1.TypeMeta `json:",inline"`
	// Name is unique
	Name string `json:"name" protobuf:"3"`
	// Project level defaults
	Spec ProjEnvCapsuleBase `json:"spec" protobuf:"4"`
}

//+kubebuilder:object:=true

type ProjEnvCapsuleBase struct {
	Files []File               `json:"files,omitempty" protobuf:"1"`
	Env   EnvironmentVariables `json:"env,omitempty" protobuf:"2"`
}

type EnvironmentSource struct {
	// Name is the name of the kubernetes object containing the environment source.
	Name string `json:"name" protobuf:"1"`
	// Kind is the kind of source, either ConfigMap or Secret.
	Kind EnvironmentSourceKind `json:"kind" protobuf:"2"`
}

type EnvironmentSourceKind string

var (
	EnvironmentSourceKindConfigMap EnvironmentSourceKind = "ConfigMap"
	EnvironmentSourceKindSecret    EnvironmentSourceKind = "Secret"
)

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type CapsuleSet struct {
	metav1.TypeMeta `json:",inline"`
	// Name,Project is unique
	Name string `json:"name" protobuf:"3"`
	// Project references an existing Project type with the given name
	// Will throw an error (in the platform) if the Project does not exist
	Project string `json:"project" protobuf:"4"`
	// Capsule-level defaults
	Spec            CapsuleSpec            `json:"spec" protobuf:"5"`
	Environments    map[string]CapsuleSpec `json:"environments" protobuf:"6"`
	EnvironmentRefs []string               `json:"environmentRefs" protobuf:"7"`
}

// +kubebuilder:object:root=true
type Capsule struct {
	metav1.TypeMeta `json:",inline"`
	// Name,Project,Environment is unique
	// Project,Name referes to an existing Capsule type with the given name and project
	// Will throw an error (in the platform) if the Capsule does not exist
	Name string `json:"name" protobuf:"3"`
	// Project references an existing Project type with the given name
	// Will throw an error (in the platform) if the Project does not exist
	Project string `json:"project" protobuf:"4"`
	// Environment references an existing Environment type with the given name
	// Will throw an error (in the platform) if the Environment does not exist
	// The environment also needs to be present in the parent Capsule
	Environment string      `json:"environment" protobuf:"5"`
	Spec        CapsuleSpec `json:"spec" protobuf:"6"`
}

// +kubebuilder:object:root=true

type CapsuleSpec struct {
	metav1.TypeMeta `json:",inline"`

	Annotations map[string]string `json:"annotations" protobuf:"11"`

	// Image specifies what image the Capsule should run.
	Image string `json:"image" protobuf:"3"`

	// Command is run as a command in the shell. If left unspecified, the
	// container will run using what is specified as ENTRYPOINT in the
	// Dockerfile.
	Command string `json:"command,omitempty" protobuf:"4"`

	// Args is a list of arguments either passed to the Command or if Command
	// is left empty the arguments will be passed to the ENTRYPOINT of the
	// docker image.
	Args []string `json:"args,omitempty" protobuf:"5" patchStrategy:"replace"`

	// Interfaces specifies the list of interfaces the the container should
	// have. Specifying interfaces will create the corresponding kubernetes
	// Services and Ingresses depending on how the interface is configured.
	// nolint:lll
	Interfaces []CapsuleInterface `json:"interfaces,omitempty" protobuf:"6" patchMergeKey:"port" patchStrategy:"merge"`

	// Files is a list of files to mount in the container. These can either be
	// based on ConfigMaps or Secrets.
	Files []File `json:"files" protobuf:"7" patchMergeKey:"path" patchStrategy:"merge"`

	// Env defines the environment variables set in the Capsule
	Env EnvironmentVariables `json:"env" protobuf:"12"`

	// Scale specifies the scaling of the Capsule.
	Scale Scale `json:"scale,omitempty" protobuf:"8"`

	CronJobs []CronJob `json:"cronJobs,omitempty" protobuf:"10" patchMergeKey:"name" patchStrategy:"replace"`

	// TODO Move to plugin
	AutoAddRigServiceAccounts bool `json:"autoAddRigServiceAccounts" protobuf:"13"`

	Extensions Extensions `json:"extensions,omitempty" protobuf:"14"`
}

// EnvironmentVariables defines the environment variables injected into a Capsule.
type EnvironmentVariables struct {
	// Raw is a list of environment variables as key-value pairs.
	Raw map[string]string `json:"raw" protobuf:"1"`
	// Sources is a list of source files which will be injected as environment variables.
	// They can be references to either ConfigMaps or Secrets.
	Sources []EnvironmentSource `json:"sources" protobuf:"2"`
}

type File struct {
	Path     string         `json:"path,omitempty" protobuf:"1"`
	AsSecret bool           `json:"asSecret,omitempty" protobuf:"3"`
	Bytes    *[]byte        `json:"bytes,omitempty" protobuf:"4"`
	String   *string        `json:"string,omitempty" protobuf:"5"`
	Ref      *FileReference `json:"ref,omitempty" protobuf:"6"`
}

// FileReference defines the name of a k8s config resource and the key from which
// to retrieve the contents
type FileReference struct {
	// Kind of reference. Can be either ConfigMap or Secret.
	Kind string `json:"kind" protobuf:"1"`

	// Name of reference.
	Name string `json:"name" protobuf:"2"`

	// Key in reference which holds file contents.
	Key string `json:"key" protobuf:"3"`
}

type Scale struct {
	// Horizontal specifies the horizontal scaling of the Capsule.
	Horizontal HorizontalScale `json:"horizontal,omitempty" protobuf:"1"`

	// Vertical specifies the vertical scaling of the Capsule.
	Vertical *VerticalScale `json:"vertical,omitempty" protobuf:"2"`
}

// HorizontalScale defines the policy for the number of replicas of
// the capsule It can both be configured with autoscaling and with a
// static number of replicas
type HorizontalScale struct {
	// Min specifies the minimum amount of instances to run.
	Min uint32 `json:"min" protobuf:"4"`

	// Max specifies the maximum amount of instances to run. Omit to
	// disable autoscaling.
	Max *uint32 `json:"max,omitempty" protobuf:"5"`

	// Instances specifies minimum and maximum amount of Capsule
	// instances.
	// Deprecated; use `min` and `max` instead.
	Instances *Instances `json:"instances,omitempty" protobuf:"1"`

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

func (i *Instances) ToK8s() *v1alpha2.Instances {
	if i == nil {
		return nil
	}
	return &v1alpha2.Instances{
		Min: i.Min,
		Max: ptr.Copy(i.Max),
	}
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

func (c *CPUTarget) ToK8s() *v1alpha2.CPUTarget {
	if c == nil {
		return nil
	}
	return &v1alpha2.CPUTarget{
		Utilization: ptr.Copy(c.Utilization),
	}
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

func (c CustomMetric) ToK8s() v1alpha2.CustomMetric {
	return v1alpha2.CustomMetric{
		InstanceMetric: c.InstanceMetric.ToK8s(),
		ObjectMetric:   c.ObjectMetric.ToK8s(),
	}
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

func (i *InstanceMetric) ToK8s() *v1alpha2.InstanceMetric {
	if i == nil {
		return nil
	}
	return &v1alpha2.InstanceMetric{
		MetricName:   i.MetricName,
		MatchLabels:  maps.Clone(i.MatchLabels),
		AverageValue: i.AverageValue,
	}
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

func (o *ObjectMetric) ToK8s() *v1alpha2.ObjectMetric {
	if o == nil {
		return nil
	}
	return &v1alpha2.ObjectMetric{
		MetricName:      o.MetricName,
		MatchLabels:     maps.Clone(o.MatchLabels),
		AverageValue:    o.AverageValue,
		Value:           o.Value,
		DescribedObject: *o.DescribedObject.DeepCopy(),
	}
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

func (v *VerticalScale) ToK8s() *v1alpha2.VerticalScale {
	if v == nil {
		return nil
	}
	return &v1alpha2.VerticalScale{
		CPU:    v.CPU.ToK8s(),
		Memory: v.Memory.ToK8s(),
		GPU:    v.GPU.ToK8s(),
	}
}

// ResourceLimits specifies the request and limit of a resource.
type ResourceLimits struct {
	// Request specifies the resource request.
	Request *resource.Quantity `json:"request,omitempty" protobuf:"1"`
	// Limit specifies the resource limit.
	Limit *resource.Quantity `json:"limit,omitempty" protobuf:"2"`
}

func (r *ResourceLimits) ToK8s() *v1alpha2.ResourceLimits {
	if r == nil {
		return nil
	}
	res := &v1alpha2.ResourceLimits{
		// Request: r.Request.DeepCopy(),
		// Limit:   r.Limit.DeepCopy(),
	}
	if r.Request != nil {
		rr := r.Request.DeepCopy()
		res.Request = &rr
	}
	if r.Limit != nil {
		l := r.Limit.DeepCopy()
		res.Limit = &l
	}
	return res
}

// ResourceRequest specifies the request of a resource.
type ResourceRequest struct {
	// Request specifies the request of a resource.
	Request resource.Quantity `json:"request,omitempty" protobuf:"1"`
}

func (r *ResourceRequest) ToK8s() *v1alpha2.ResourceRequest {
	if r == nil {
		return nil
	}
	return &v1alpha2.ResourceRequest{
		Request: r.Request.DeepCopy(),
	}
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

	// Host routes that are mapped to this interface.
	Routes []HostRoute `json:"routes,omitempty" protobuf:"6"`
}

func (c CapsuleInterface) ToK8s() v1alpha2.CapsuleInterface {
	res := v1alpha2.CapsuleInterface{
		Name:      c.Name,
		Port:      c.Port,
		Liveness:  c.Liveness.ToK8s(),
		Readiness: c.Readiness.ToK8s(),
		Routes:    []v1alpha2.HostRoute{},
	}
	for _, r := range c.Routes {
		res.Routes = append(res.Routes, r.ToK8s())
	}
	return res
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

func (h HostRoute) ToK8s() v1alpha2.HostRoute {
	res := v1alpha2.HostRoute{
		ID:           h.ID,
		Host:         h.Host,
		RouteOptions: h.RouteOptions.ToK8s(),
	}
	for _, p := range h.Paths {
		res.Paths = append(res.Paths, p.ToK8s())
	}
	return res
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

func (h HTTPPathRoute) ToK8s() v1alpha2.HTTPPathRoute {
	return v1alpha2.HTTPPathRoute{
		Path:  h.Path,
		Match: v1alpha2.PathMatchType(h.Match),
	}
}

// Route options.
type RouteOptions struct {
	// Annotations of the route option. This can be plugin-specific configuration
	// that allows custom plugins to add non-standard behavior.
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"4"`
}

func (r RouteOptions) ToK8s() v1alpha2.RouteOptions {
	return v1alpha2.RouteOptions{
		Annotations: maps.Clone(r.Annotations),
	}
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

func (i *InterfaceLivenessProbe) ToK8s() *v1alpha2.InterfaceLivenessProbe {
	if i == nil {
		return nil
	}
	return &v1alpha2.InterfaceLivenessProbe{
		Path:         i.Path,
		TCP:          i.TCP,
		GRPC:         i.GRPC.ToK8s(),
		StartupDelay: i.StartupDelay,
	}
}

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

func (i *InterfaceReadinessProbe) ToK8s() *v1alpha2.InterfaceReadinessProbe {
	if i == nil {
		return nil
	}
	return &v1alpha2.InterfaceReadinessProbe{
		Path: i.Path,
		TCP:  i.TCP,
		GRPC: i.GRPC.ToK8s(),
	}
}

// InterfaceGRPCProbe specifies a GRPC probe.
type InterfaceGRPCProbe struct {
	// Service specifies the gRPC health probe service to probe. This is a
	// used as service name as per standard gRPC health/v1.
	Service string `json:"service" protobuf:"1"`

	// Enabled controls if the gRPC health check is activated.
	Enabled bool `json:"enabled,omitempty" protobuf:"2"`
}

func (i *InterfaceGRPCProbe) ToK8s() *v1alpha2.InterfaceGRPCProbe {
	if i == nil {
		return nil
	}
	return &v1alpha2.InterfaceGRPCProbe{
		Service: i.Service,
		Enabled: i.Enabled,
	}
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

func (c CronJob) ToK8s() v1alpha2.CronJob {
	return v1alpha2.CronJob{
		Name:           c.Name,
		Schedule:       c.Schedule,
		URL:            c.URL.ToK8s(),
		Command:        c.Command.ToK8s(),
		MaxRetries:     ptr.Copy(c.MaxRetries),
		TimeoutSeconds: ptr.Copy(c.TimeoutSeconds),
	}
}

type URL struct {
	// +kubebuilder:validation:Required
	Port uint16 `json:"port" protobuf:"1"`
	// +kubebuilder:validation:Required
	Path            string            `json:"path" protobuf:"2"`
	QueryParameters map[string]string `json:"queryParameters,omitempty" protobuf:"3"`
}

func (u *URL) ToK8s() *v1alpha2.URL {
	if u == nil {
		return nil
	}
	return &v1alpha2.URL{
		Port:            u.Port,
		Path:            u.Path,
		QueryParameters: maps.Clone(u.QueryParameters),
	}
}

type JobCommand struct {
	// +kubebuilder:validation:Required
	Command string   `json:"command" protobuf:"1"`
	Args    []string `json:"args,omitempty" protobuf:"2"`
}

func (j *JobCommand) ToK8s() *v1alpha2.JobCommand {
	if j == nil {
		return nil
	}
	return &v1alpha2.JobCommand{
		Command: j.Command,
		Args:    slices.Clone(j.Args),
	}
}

type Extensions struct {
	Fields map[string]string `json:"fields,omitempty" protobuf:"1"`
}

func (e Extensions) ToK8s() v1alpha2.Extensions {
	return v1alpha2.Extensions{
		Fields: maps.Clone(e.Fields),
	}
}

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "platform.rig.dev", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&Capsule{}, &Project{}, &Environment{}, &CapsuleSet{})
}
