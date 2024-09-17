package v1alpha1

import (
	"github.com/rigdev/rig/pkg/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	SchemeBuilder.Register(&OperatorConfig{})
}

// OperatorConfig is the Schema for the operator config API
// +kubebuilder:object:root=true
type OperatorConfig struct {
	metav1.TypeMeta `json:",inline"`

	// WebhooksEnabled sets wether or not webhooks should be enabled. When
	// enabled a certificate should be mounted at the webhook server
	// certificate path. Defaults to true if omitted.
	WebhooksEnabled *bool `json:"webhooksEnabled,omitempty"`

	// DevModeEnabled enables verbose logs and changes the logging format to be
	// more human readable.
	DevModeEnabled bool `json:"devModeEnabled,omitempty"`

	// LeaderElectionEnabled enables leader election when running multiple
	// instances of the operator.
	LeaderElectionEnabled *bool `json:"leaderElectionEnabled,omitempty"`

	// Pipeline defines the capsule controller pipeline
	Pipeline Pipeline `json:"pipeline,omitempty"`
}

type Pipeline struct {
	// How to handle the service account step of capsules in the cluster.
	// Defaults to rigdev.service_account.
	ServiceAccountStep CapsuleStep `json:"serviceAccountStep,omitempty"`
	// How to handle the deployment step of capsules in the cluster.
	// Defaults to rigdev.deployment.
	DeploymentStep CapsuleStep `json:"deploymentStep,omitempty"`
	// How to handle the routes for capsules in the cluster.
	// If left empty, routes will not be handled.
	RoutesStep CapsuleStep `json:"routesStep,omitempty"`
	// How to handle the cronjob step of capsules in the cluster.
	// Defaults to rigdev.cron_jobs
	CronJobsStep CapsuleStep `json:"cronJobsStep,omitempty"`
	// How to handle the VPA step of capsules in the cluster.
	// If left empty, no VPAs will be created.
	VPAStep CapsuleStep `json:"vpaStep,omitempty"`
	// How to handle the service monitor step of capsules in the cluster.
	// If left empty, no service monitors will be created.
	// rigdev.service_monitor plugin spawns a Prometheus ServiceMonitor per capsule
	// for use with a Prometheus Operator stack.
	ServiceMonitorStep CapsuleStep `json:"serviceMonitorStep,omitempty"`
	// Steps to perform as part of running the operator.
	// +patchStrategy=merge
	Steps []Step `json:"steps,omitempty"`
	// CustomPlugins enables custom plugins to be injected into the
	// operator. The plugins injected here can then be referenced in 'steps'
	CustomPlugins []CustomPlugin `json:"customPlugins,omitempty"`
	// CapsuleExtensions supported by the Operator. Each extension supported
	// should be configured in the map, with an additional plugin name.
	CapsuleExtensions map[string]CapsuleStep `json:"capsuleExtensions,omitempty"`
}

type CapsuleStep struct {
	// The plugin to use for handling the capsule step.
	// fx. "rigdev.ingress_routes" for routesStep will create an ingress resource per route.
	// fx. "rigdev.deployment" for deploymentStep will use the default deployment plugin.
	Plugin string `json:"plugin,omitempty"`

	// Config is a string defining the plugin-specific configuration of the plugin.
	Config string `json:"config,omitempty"`
}

type Step struct {
	// Optional tag which is readable by plugins when executed
	Tag string `json:"tag,omitempty"`
	// Match requirements for running the Step on a given Capsule.
	Match CapsuleMatch `json:"match,omitempty"`

	// Plugins to run as part of this step.
	Plugins []Plugin `json:"plugins,omitempty"`

	// If set, only capsules in one of the namespaces given will have this step run.
	// Deprecated, use Match.Namespaces.
	Namespaces []string `json:"namespaces,omitempty"`
	// If set, only execute the plugin on the capsules specified.
	// Deprecated, use Match.Names.
	Capsules []string `json:"capsules,omitempty"`
	// If set, will enable the step for the Rig platform which is a Capsule as well
	// Deprecated, use Match.EnableForPlatform.
	EnableForPlatform bool `json:"enableForPlatform,omitempty"`
}

type CapsuleMatch struct {
	// If set, only capsules in one of the namespaces given will have this step run.
	Namespaces []string `json:"namespaces,omitempty"`
	// If set, only execute the plugin on the capsules specified.
	Names []string `json:"names,omitempty"`
	// If set, only execute the plugin on the capsules matching the annotations.
	Annotations map[string]string `json:"annotations,omitempty"`
	// If set, will enable the step for the Rig platform which is a Capsule as well
	EnableForPlatform bool `json:"enableForPlatform,omitempty"`
}

type CustomPlugin struct {
	// The container image which supplies the plugins
	Image string `json:"image"`
}

type Plugin struct {
	// Optional tag which is readable by plugin when executed
	Tag string `json:"tag,omitempty"`
	// Name of the plugin to run.
	// Deprecated, use Plugin.
	Name string `json:"name,omitempty"`
	// Name of the plugin to run.
	Plugin string `json:"plugin,omitempty"`
	// Config is a string defining the plugin-specific configuration of the plugin.
	Config string `json:"config,omitempty"`
}

func (p Plugin) GetPlugin() string {
	if p.Plugin != "" {
		return p.Plugin
	}

	return p.Name
}

type VerticalPodAutoscaler struct {
	// Enabled enables the creation of a VerticalPodAutoscaler per capsule
	Enabled bool `json:"enabled,omitempty"`
}

type PrometheusServiceMonitor struct {
	// Path is the path which Prometheus should query on ports. Defaults to /metrics if not set.
	Path string `json:"path,omitempty"`
	// PortName is the name of the port which Prometheus will query metrics on
	PortName string `json:"portName"`
}

func (c *OperatorConfig) Default() *OperatorConfig {
	if c == nil {
		return c
	}
	c.SetGroupVersionKind(schema.FromAPIVersionAndKind(
		GroupVersion.Identifier(),
		"OperatorConfig",
	))
	if c.WebhooksEnabled == nil {
		c.WebhooksEnabled = ptr.New(true)
	}
	if c.LeaderElectionEnabled == nil {
		c.LeaderElectionEnabled = ptr.New(true)
	}
	return c
}
