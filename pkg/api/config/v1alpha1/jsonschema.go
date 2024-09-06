package v1alpha1

// JSONSchema is a JSON-Schema of a subset of Specification Draft 4 (http://json-schema.org/).
type JSONSchema struct {
	Description      string   `json:"description,omitempty"`
	Type             string   `json:"type,omitempty"`
	Nullable         bool     `json:"nullable,omitempty"`
	Format           string   `json:"format,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMaximum bool     `json:"exclusiveMaximum,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty"`
	ExclusiveMinimum bool     `json:"exclusiveMinimum,omitempty"`
	MaxLength        *int64   `json:"maxLength,omitempty"`
	MinLength        *int64   `json:"minLength,omitempty"`
	Enum             []string `json:"enum,omitempty"`
}

type JSONObjectSchema struct {
	Properties map[string]JSONSchema `json:"properties,omitempty"`
}
