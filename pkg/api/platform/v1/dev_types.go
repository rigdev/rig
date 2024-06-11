// +kubebuilder:object:generate=true
// +groupName=rig.platform
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type HostCapsule struct {
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
	Environment string `json:"environment" protobuf:"5"`

	// Network mapping between the host network and the Kubernetes cluster network. When activated,
	// traffic between the two networks will be tunneled according to the rules specified here.
	Network HostNetwork `json:"network" protobuf:"6"`
}

type HostNetwork struct {
	// HostInterfaces are interfaces activated on the local machine (the host) and forwarded
	// to the Kubernetes cluster capsules.
	HostInterfaces []HostInterface `json:"hostInterfaces" protobuf:"1"`

	// CapsuleInterfaces are interfaces activated on the Capsule within the Kubernetes cluster
	// and forwarded to the local machine (the host). The traffic is directed to a single target,
	// e.g. `localhost:8080`.
	CapsuleInterface []CapsuleInterface `json:"capsuleInterfaces" protobuf:"2"`

	// TunnelPort for which the proxy-capsule should listen on. This is automatically set by the tooling.
	TunnelPort uint32 `json:"tunnelPort,omitempty" protobuf:"3"`
}

type HostInterface struct {
	// Port on the host from where to accept traffic from.
	Port *uint32 `json:"port,omitempty" protobuf:"2"`
	// CapsuleTarget is the capsule-name:capsule-port to forward traffic to.
	CapsuleTarget string           `json:"capsuleTarget" protobuf:"1"`
	Options       InterfaceOptions `json:"options,omitempty" protobuf:"3"`
}

type CapsuleInterface struct {
	// Port on the Capsule from where to accept traffic from.
	Port uint32 `json:"port" protobuf:"1"`
	// HostTarget is the local address:port to forward traffic to.
	HostTarget string           `json:"hostTarget,omitempty" protobuf:"2"`
	Options    InterfaceOptions `json:"options,omitempty" protobuf:"3"`
}

type InterfaceOptions struct {
	// TCP enables layer-4 proxying in favor of layer-7 HTTP proxying.
	TCP bool `json:"tcp,omitempty" protobuf:"1"`
	// AllowOrigin sets the `Access-Control-Allow-Origin` Header on responses to
	// the provided value, allowing local by-pass of CORS rules.
	// Ignored if TCP is enabled.
	AllowOrigin string `json:"allowOrigin,omitempty" protobuf:"2"`
	// ChangeOrigin changes the Host header to match the given target. If not set,
	// the Host header will be that of the original request.
	// This does not impact the Origin header - use `Headers` to set that.
	// Ignored if TCP is enabled.
	ChangeOrigin bool `json:"changeOrigin,omitempty" protobuf:"3"`
	// Headers to set on the proxy-requests.
	// Ignored if TCP is enabled.
	Headers map[string]string `json:"headers,omitempty" protobuf:"4"`
}
