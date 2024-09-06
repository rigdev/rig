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
	// Enums are annoying with our API doc generation.
	// Properly, an Enum is a []any but our API doc generation cannot generate docs for 'any'
	// Even if we would restrict Enums to all be of just one type, we still have the issue of having multiple type offerings for Enum
	// E.g.
	//   EnumString []string
	//   EnumInt []int
	// since to be compatible with JSONSchema both fields must map to an `enum` json field.
	Enum []string `json:"enum,omitempty"`
}
