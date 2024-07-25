package flags

import (
	"os"

	"github.com/rigdev/rig/cmd/common"
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

var Flags FlagsStruct

func InitFlags() {
	Flags = FlagsStruct{
		OutputType:     common.OutputTypePretty,
		NonInteractive: false,
		Environment:    os.Getenv("RIG_ENVIRONMENT"),
		Project:        os.Getenv("RIG_PROJECT"),
		BasicAuth:      false,
		Host:           os.Getenv("RIG_HOST"),
		Context:        "",
	}
}
