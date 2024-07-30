// +kubebuilder:object:generate=true
// +groupName=rig.platform
package v1

import (
	"github.com/rigdev/rig/pkg/api/v1alpha2"
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
	Interfaces []v1alpha2.CapsuleInterface `json:"interfaces,omitempty" protobuf:"6" patchMergeKey:"port" patchStrategy:"merge"`

	// Files is a list of files to mount in the container. These can either be
	// based on ConfigMaps or Secrets.
	Files []File `json:"files" protobuf:"7" patchMergeKey:"path" patchStrategy:"merge"`

	// Env defines the environment variables set in the Capsule
	Env EnvironmentVariables `json:"env" protobuf:"12"`

	// Scale specifies the scaling of the Capsule.
	Scale Scale `json:"scale,omitempty" protobuf:"8"`

	CronJobs []v1alpha2.CronJob `json:"cronJobs,omitempty" protobuf:"10" patchMergeKey:"name" patchStrategy:"replace"`

	// TODO Move to plugin
	AutoAddRigServiceAccounts bool `json:"autoAddRigServiceAccounts" protobuf:"13"`
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
	Path     string  `json:"path,omitempty" protobuf:"1"`
	AsSecret bool    `json:"asSecret,omitempty" protobuf:"3"`
	Bytes    *[]byte `json:"bytes,omitempty" protobuf:"4"`
	String   *string `json:"string,omitempty" protobuf:"5"`
	// TODO Ref
}

type Scale struct {
	// Horizontal specifies the horizontal scaling of the Capsule.
	Horizontal HorizontalScale `json:"horizontal,omitempty" protobuf:"1"`

	// Vertical specifies the vertical scaling of the Capsule.
	Vertical *v1alpha2.VerticalScale `json:"vertical,omitempty" protobuf:"2"`
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
	Instances *v1alpha2.Instances `json:"instances,omitempty" protobuf:"1"`

	// CPUTarget specifies that this Capsule should be scaled using CPU
	// utilization.
	CPUTarget *v1alpha2.CPUTarget `json:"cpuTarget,omitempty" protobuf:"2"`
	// CustomMetrics specifies custom metrics emitted by the custom.metrics.k8s.io API
	// which the autoscaler should scale on
	CustomMetrics []v1alpha2.CustomMetric `json:"customMetrics,omitempty" protobuf:"3" patchStrategy:"replace"`
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
