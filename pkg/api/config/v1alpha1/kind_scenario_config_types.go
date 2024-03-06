package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KindScenarioConfig is the Schema for the CLI kind example config
// +kubebuilder:object:root=true
type KindScenarioConfig struct {
	metav1.TypeMeta `json:",inline"`

	// Name of scenario
	Name string `json:"name"`

	// Steps is a list of steps to perform against the cluster after creation
	Steps []*KindScenarioStep `json:"steps,omitempty"`

	// KindConfig is the kind configuration we will use to create the kind
	// cluster. If left empty we will use the default rig example
	// configuration.
	KindConfig string `json:"kindConfig,omitempty"`
}

type KindScenarioStep struct {
	Helm     *KindScenarioHelmStep     `json:"helm,omitempty"`
	Manifest *KindScenarioManifestStep `json:"manifest,omitempty"`
	Exec     *KindScenarioExecStep     `json:"exec,omitempty"`
}

type KindScenarioHelmStep struct {
	// Name is the helm release name.
	Name string `json:"name"`

	// Namespace is the namespace where the helm release will be installed.
	Namespace string `json:"namespace"`

	// Chart is the name of the helm chart.
	Chart string `json:"chart"`

	// ChartEnvVar specifies an environment variable which will be used instead
	// of `Chart` if the environment variable is non-empty.
	ChartEnvVar string `json:"chartEnvVar,omitempty"`

	// Repo is a URL pointing to the helm repository where the chart lives.
	Repo string `json:"repo,omitempty"`

	// Version is the chart version to install
	Version string `json:"version,omitempty"`

	// ValuesFromEnvVars specifies a mapping between values which will be set
	// to the contents of the environment variable if it is non-empty. For the
	// keys we use the same syntax as when using `--set` in the helm CLI.
	ValuesFromEnvVars map[string]string `json:"valuesFromEnvVars,omitempty"`

	// ValueFiles is a list of paths relative from scenario.yaml to a files
	// holding values for the helm release.
	ValueFiles []string `json:"valueFiles,omitempty"`

	// Values specifies helm values
	Values map[string]interface{} `json:"values,omitempty"`

	// Wait toggle wether or not to set the `--wait` flag for helm.
	Wait bool `json:"wait,omitempty"`
}

type KindScenarioManifestStep struct {
	// Path is the relative path to the manifest file or folder to apply in the
	// step.
	Path string `json:"path"`
}

type KindScenarioExecStep struct {
	// Reference to pod where command should be executed. This can either
	// be a pod name or a TYPE/name eg. deployment/mydeploy.
	Reference string `json:"reference"`

	// Namespace where pod lives
	Namespace string `json:"namespace"`

	// Container is the name of the container where the command should be run
	Container string `json:"container,omitempty"`

	// TTY wether to use tty
	TTY bool `json:"tty,omitempty"`

	// Stdin wether to attach stdin
	Stdin bool `json:"stdin,omitempty"`

	// Command is the command to run
	Command []string `json:"command"`
}

func init() {
	SchemeBuilder.Register(&KindScenarioConfig{})
}
