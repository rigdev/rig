package base

import "github.com/rigdev/rig/cmd/common"

type FlagsStruct struct {
	KubeContext string
	KubeConfig  string
	RigConfig   string
	RigContext  string
	Namespace   string
	Project     string
	Environment string
	KubeFile    string
	OutputType  common.OutputType
}

var Flags = FlagsStruct{
	KubeContext: "",
	KubeConfig:  "",
	Namespace:   "",
	RigConfig:   "",
	RigContext:  "",
	Environment: "",
	OutputType:  common.OutputTypePretty,
}
