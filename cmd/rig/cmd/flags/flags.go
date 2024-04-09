package flags

import (
	"os"

	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/cli/scope"
)

//nolint:revive
type FlagsStruct struct {
	OutputType     common.OutputType
	NonInteractive bool
	Environment    string
	Project        string
	BasicAuth      bool
	Host           string
	Context        string
}

var Flags = FlagsStruct{
	OutputType:     common.OutputTypePretty,
	NonInteractive: false,
	Environment:    os.Getenv("RIG_ENVIRONMENT"),
	Project:        os.Getenv("RIG_PROJECT"),
	BasicAuth:      false,
	Host:           os.Getenv("RIG_HOST"),
	Context:        "",
}

func GetEnvironment(scope scope.Scope) string {
	if Flags.Environment != "" {
		return Flags.Environment
	}
	return scope.GetCurrentContext().EnvironmentID
}

func GetProject(scope scope.Scope) string {
	if Flags.Project != "" {
		return Flags.Project
	}
	return scope.GetCurrentContext().ProjectID
}
