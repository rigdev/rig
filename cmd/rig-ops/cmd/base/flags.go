package base

import "github.com/rigdev/rig/cmd/common"

type FlagsStruct struct {
	KubeContext    string
	KubeConfig     string
	RigConfig      string
	RigContext     string
	Namespace      string
	Project        string
	Environment    string
	KubeFile       string
	OperatorConfig string
	OutputType     common.OutputType
}

var Flags = FlagsStruct{
	OutputType: common.OutputTypePretty,
}
