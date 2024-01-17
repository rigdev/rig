package base

import (
	"errors"

	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
)

type OutputType string

const (
	OutputTypeJSON   OutputType = "json"
	OutputTypeYAML   OutputType = "yaml"
	OutputTypePretty OutputType = "pretty"
)

// String is used both by fmt.Print and by Cobra in help text
func (e *OutputType) String() string {
	return string(*e)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *OutputType) Set(v string) error {
	switch v {
	case string(OutputTypeJSON), string(OutputTypeYAML), string(OutputTypePretty):
		*e = OutputType(v)
		return nil
	default:
		return errors.New(`must be one of "json", "yaml", or "pretty"`)
	}
}

// Type is only used in help text
func (e *OutputType) Type() string {
	return "OutputType"
}

type FlagsStruct struct {
	OutputType     OutputType
	NonInteractive bool
	Environment    string
}

var Flags = FlagsStruct{
	OutputType:     OutputTypePretty,
	NonInteractive: false,
	Environment:    "",
}

func GetEnvironment(cfg *cmdconfig.Config) string {
	if Flags.Environment != "" {
		return Flags.Environment
	}
	return cfg.GetEnvironment()
}
