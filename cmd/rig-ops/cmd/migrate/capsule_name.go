package migrate

import "github.com/rigdev/rig/pkg/errors"

type CapsuleName string

// String is used both by fmt.Print and by Cobra in help text
func (c *CapsuleName) String() string {
	return string(*c)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (c *CapsuleName) Set(v string) error {
	switch v {
	case string(CapsuleNameService), string(CapsuleNameDeployment), string(CapsuleNameInput):
		*c = CapsuleName(v)
		return nil
	default:
		return errors.New(`must be one of "service", "deployment", or "input"`)
	}
}

// Type is only used in help text
func (c *CapsuleName) Type() string {
	return "string"
}

const (
	CapsuleNameService    CapsuleName = "service"
	CapsuleNameDeployment CapsuleName = "deployment"
	CapsuleNameInput      CapsuleName = "input"
)
