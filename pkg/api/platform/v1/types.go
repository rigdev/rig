package v1

import (
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProjEnvCapsuleBase struct {
	ConfigFiles          []ConfigFile      `json:"configFiles,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
}

type EnvironmentSource struct {
	Name string                `json:"name"`
	Kind EnvironmentSourceKind `json:"kind"`
}

type EnvironmentSourceKind string

var (
	EnvironmentSourceKindConfigMap EnvironmentSourceKind = "config_map"
	EnvironmentSourceKindSecret    EnvironmentSourceKind = "secret"
)

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type Environment struct {
	metav1.TypeMeta `json:",inline"`
	// Name is unique
	Name              string `json:"name"`
	NamespaceTemplate string `json:"namespaceTemplate"`
	OperatorVersion   string `json:"operatorVersion"`
	ClusterID         string `json:"clusterID"`
	// Environment level defaults
	CapsuleBase ProjEnvCapsuleBase `json:"capsuleBase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type Project struct {
	metav1.TypeMeta `json:",inline"`
	// Name is unique
	Name string `json:"name"`
	// A capsule is only allowed in an environment if its project references the environment in this list
	Environments []string `json:"environments"`
	// Project level defaults
	CapsuleBase ProjEnvCapsuleBase `json:"capsuleBase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

type CapsuleStar struct {
	metav1.TypeMeta `json:",inline"`
	// Name,Project is unique
	Name string `json:"name"`
	// Project references an existing Project2 type with the given name
	// Will throw an error (in the platform) if the project does not exist
	Project string `json:"project"`
	// Capsule-level defaults
	CapsuleBase  CapsuleSpecExtension            `json:"capsuleBase"`
	Environments map[string]CapsuleSpecExtension `json:"environments"`
}

type CapsuleSpecExtension struct {
	metav1.TypeMeta `json:",inline"`
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
	Interfaces []v1alpha2.CapsuleInterface `json:"interfaces,omitempty"`

	// Files is a list of files to mount in the container. These can either be
	// based on ConfigMaps or Secrets.
	ConfigFiles []ConfigFile `json:"configFiles"`

	// Scale specifies the scaling of the Capsule.
	Scale v1alpha2.CapsuleScale `json:"scale,omitempty"`

	// NodeSelector is a selector for what nodes the Capsule should live on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	CronJobs []v1alpha2.CronJob `json:"cronJobs,omitempty"`

	Annotations map[string]string `json:"annotations"`
}

type ConfigFile struct {
	Path     string `json:"path,omitempty"`
	Content  []byte `json:"content,omitempty"`
	IsSecret bool   `json:"isSecret,omitempty"`
}
