// +kubebuilder:object:generate=true
// +groupName=rig.platform
package v1

import (
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type Environment struct {
	metav1.TypeMeta `json:",inline"`
	// Name is unique
	Name              string `json:"name" protobuf:"3"`
	NamespaceTemplate string `json:"namespaceTemplate" protobuf:"4"`
	OperatorVersion   string `json:"operatorVersion" protobuf:"5"`
	ClusterID         string `json:"clusterID" protobuf:"6"`
	// Environment level defaults
	CapsuleBase ProjEnvCapsuleBase `json:"capsuleBase" protobuf:"7"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type Project struct {
	metav1.TypeMeta `json:",inline"`
	// Name is unique
	Name string `json:"name" protobuf:"3"`
	// A capsule is only allowed in an environment if its project references the environment in this list
	Environments []string `json:"environments" protobuf:"4"`
	// Project level defaults
	CapsuleBase ProjEnvCapsuleBase `json:"capsuleBase" protobuf:"5"`
}

//+kubebuilder:object:=true

type ProjEnvCapsuleBase struct {
	ConfigFiles          []ConfigFile      `json:"configFiles,omitempty" protobuf:"1"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty" protobuf:"2"`
}

type EnvironmentSource struct {
	Name string                `json:"name" protobuf:"1"`
	Kind EnvironmentSourceKind `json:"kind" protobuf:"2"`
}

type EnvironmentSourceKind string

var (
	EnvironmentSourceKindConfigMap EnvironmentSourceKind = "config_map"
	EnvironmentSourceKindSecret    EnvironmentSourceKind = "secret"
)

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type CapsuleStar struct {
	metav1.TypeMeta `json:",inline"`
	// Name,Project is unique
	Name string `json:"name" protobuf:"3"`
	// Project references an existing Project type with the given name
	// Will throw an error (in the platform) if the Project does not exist
	Project string `json:"project" protobuf:"4"`
	// Capsule-level defaults
	CapsuleBase  CapsuleSpecExtension `json:"capsuleBase" protobuf:"5"`
	Environments []string             `json:"environments" protobuf:"6"`
}

// +kubebuilder:object:root=true

type CapsuleEnvironment struct {
	metav1.TypeMeta `json:",inline"`
	// Name,Project,Environment is unique
	// Project,Name referes to an existing CapsuleStar type with the given name and project
	// Will throw an error (in the platform) if the CapsuleStar does not exist
	Name string `json:"name" protobuf:"3"`
	// Project references an existing Project type with the given name
	// Will throw an error (in the platform) if the Project does not exist
	Project string `json:"project" protobuf:"4"`
	// Environment references an existing Environment type with the given name
	// Will throw an error (in the platform) if the Environment does not exist
	// The environment also needs to be present in the parent CapsuleStar
	Environment string               `json:"environment" protobuf:"5"`
	Spec        CapsuleSpecExtension `json:"spec" protobuf:"6"`
}

type CapsuleSpecExtension struct {
	metav1.TypeMeta `json:",inline"`
	// Image specifies what image the Capsule should run.
	Image string `json:"image" protobuf:"3"`

	// Command is run as a command in the shell. If left unspecified, the
	// container will run using what is specified as ENTRYPOINT in the
	// Dockerfile.
	Command string `json:"command,omitempty" protobuf:"4"`

	// Args is a list of arguments either passed to the Command or if Command
	// is left empty the arguments will be passed to the ENTRYPOINT of the
	// docker image.
	Args []string `json:"args,omitempty" protobuf:"5"`

	// Interfaces specifies the list of interfaces the the container should
	// have. Specifying interfaces will create the corresponding kubernetes
	// Services and Ingresses depending on how the interface is configured.
	Interfaces []v1alpha2.CapsuleInterface `json:"interfaces,omitempty" protobuf:"6"`

	// Files is a list of files to mount in the container. These can either be
	// based on ConfigMaps or Secrets.
	ConfigFiles []ConfigFile `json:"configFiles" protobuf:"7"`

	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty" protobuf:"12"`

	// Scale specifies the scaling of the Capsule.
	Scale v1alpha2.CapsuleScale `json:"scale,omitempty" protobuf:"8"`

	// NodeSelector is a selector for what nodes the Capsule should live on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"9"`

	CronJobs []v1alpha2.CronJob `json:"cronJobs,omitempty" protobuf:"10"`

	Annotations map[string]string `json:"annotations" protobuf:"11"`
}

type ConfigFile struct {
	Path     string `json:"path,omitempty" protobuf:"1"`
	Content  []byte `json:"content,omitempty" protobuf:"2"`
	IsSecret bool   `json:"isSecret,omitempty" protobuf:"3"`
}
