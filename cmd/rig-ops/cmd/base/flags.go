package base

type FlagsStruct struct {
	KubeContext string
	KubeConfig  string
	RigConfig   string
	RigContext  string
	Namespace   string
	Project     string
	Environment string
	KubeFile    string
}

var Flags = FlagsStruct{
	KubeContext: "",
	KubeConfig:  "",
	Namespace:   "",
	RigConfig:   "",
	RigContext:  "",
	Project:     "",
	Environment: "",
}
