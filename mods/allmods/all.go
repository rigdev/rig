package allmods

import (
	"github.com/rigdev/rig/mods/annotations"
	"github.com/rigdev/rig/mods/datadog"
	envmapping "github.com/rigdev/rig/mods/env_mapping"
	googlesqlproxy "github.com/rigdev/rig/mods/google_cloud_sql_auth_proxy"
	ingressroutes "github.com/rigdev/rig/mods/ingress_routes"
	initcontainer "github.com/rigdev/rig/mods/init_container"
	objecttemplate "github.com/rigdev/rig/mods/object_template"
	"github.com/rigdev/rig/mods/placement"
	"github.com/rigdev/rig/mods/sidecar"
	"github.com/rigdev/rig/pkg/controller/mod"
)

var Mods = map[string]mod.Mod{
	annotations.Name:    &annotations.Plugin{},
	datadog.Name:        &datadog.Plugin{},
	envmapping.Name:     &envmapping.Plugin{},
	googlesqlproxy.Name: &googlesqlproxy.Plugin{},
	initcontainer.Name:  &initcontainer.Plugin{},
	objecttemplate.Name: &objecttemplate.Plugin{},
	placement.Name:      &placement.Plugin{},
	sidecar.Name:        &sidecar.Plugin{},
	ingressroutes.Name:  &ingressroutes.Plugin{},
}
