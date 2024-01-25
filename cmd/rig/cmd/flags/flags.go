package flags

import (
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
)

//nolint:revive
type FlagsStruct struct {
	OutputType     common.OutputType
	NonInteractive bool
	Environment    string
	Project        string
	BasicAuth      bool
	Host           string
}

var Flags = FlagsStruct{
	OutputType:     common.OutputTypePretty,
	NonInteractive: false,
	Environment:    "",
	Project:        "",
	BasicAuth:      false,
	Host:           "",
}

func GetEnvironment(cfg *cmdconfig.Config) string {
	if Flags.Environment != "" {
		return Flags.Environment
	}
	return cfg.GetEnvironment()
}

func GetProject(cfg *cmdconfig.Config) string {
	if Flags.Project != "" {
		return Flags.Project
	}
	return cfg.GetProject()
}
