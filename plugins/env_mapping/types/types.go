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
	// Mappings ENV:KEY
	Mappings map[string]string `json:"mappings"`
}
