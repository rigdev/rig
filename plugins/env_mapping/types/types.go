package types

const (
	AnnotationEnvMapping = "plugin.rig.dev/env-mapping"
)

type AnnotationValue struct {
	Sources []AnnotationSource `json:"sources"`
}

type AnnotationSource struct {
	// Container name default to capsule name.
	Container string `json:"container,omitempty"`
	// Optional ConfigMap reference.
	ConfigMap string `json:"configMap,omitempty"`
	// Optional Secret reference.
	Secret string `json:"secret,omitempty"`
	// Mappings within this ConfigMap or Secret.
	Mappings []AnnotationMappings `json:"mappings"`
}

type AnnotationMappings struct {
	// Env is the environment name the property should be exposed as.
	Env string `json:"env"`
	// Key is the ConfigMap or Secret property that should me mapped from.
	Key string `json:"key"`
}
