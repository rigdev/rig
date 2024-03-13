package allplugins

import (
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/plugins/annotations"
	"github.com/rigdev/rig/plugins/datadog"
	envmapping "github.com/rigdev/rig/plugins/env_mapping"
	googlesqlproxy "github.com/rigdev/rig/plugins/google_cloud_sql_auth_proxy"
	initcontainer "github.com/rigdev/rig/plugins/init_container"
	objecttemplate "github.com/rigdev/rig/plugins/object_template"
	"github.com/rigdev/rig/plugins/placement"
	"github.com/rigdev/rig/plugins/sidecar"
)

var Plugins = map[string]plugin.Server{
	annotations.Name:    &annotations.Plugin{},
	datadog.Name:        &datadog.Plugin{},
	envmapping.Name:     &envmapping.Plugin{},
	googlesqlproxy.Name: &googlesqlproxy.Plugin{},
	initcontainer.Name:  &initcontainer.Plugin{},
	objecttemplate.Name: &objecttemplate.Plugin{},
	placement.Name:      &placement.Plugin{},
	sidecar.Name:        &sidecar.Plugin{},
}
