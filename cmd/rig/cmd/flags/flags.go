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
}

var Flags = FlagsStruct{
	OutputType:     common.OutputTypePretty,
	NonInteractive: false,
	Environment:    "",
}

func GetEnvironment(cfg *cmdconfig.Config) string {
	if Flags.Environment != "" {
		return Flags.Environment
	}
	return cfg.GetEnvironment()
}
