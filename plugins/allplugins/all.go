package allplugins

import (
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/plugins/annotations"
	"github.com/rigdev/rig/plugins/cron_jobs"
	"github.com/rigdev/rig/plugins/datadog"
	"github.com/rigdev/rig/plugins/deployment"
	envmapping "github.com/rigdev/rig/plugins/env_mapping"
	googlesqlproxy "github.com/rigdev/rig/plugins/google_cloud_sql_auth_proxy"
	ingressroutes "github.com/rigdev/rig/plugins/ingress_routes"
	initcontainer "github.com/rigdev/rig/plugins/init_container"
	objecttemplate "github.com/rigdev/rig/plugins/object_template"
	"github.com/rigdev/rig/plugins/placement"
	"github.com/rigdev/rig/plugins/service_account"
	"github.com/rigdev/rig/plugins/service_monitor"
	"github.com/rigdev/rig/plugins/sidecar"
	"github.com/rigdev/rig/plugins/vpa"
)

var Plugins = map[string]plugin.Plugin{
	annotations.Name:     &annotations.Plugin{},
	datadog.Name:         &datadog.Plugin{},
	envmapping.Name:      &envmapping.Plugin{},
	googlesqlproxy.Name:  &googlesqlproxy.Plugin{},
	initcontainer.Name:   &initcontainer.Plugin{},
	objecttemplate.Name:  &objecttemplate.Plugin{},
	placement.Name:       &placement.Plugin{},
	sidecar.Name:         &sidecar.Plugin{},
	ingressroutes.Name:   &ingressroutes.Plugin{},
	deployment.Name:      &deployment.Plugin{},
	cron_jobs.Name:       &cron_jobs.Plugin{},
	service_account.Name: &service_account.Plugin{},
	service_monitor.Name: &service_monitor.Plugin{},
	vpa.Name:             &vpa.Plugin{},
}
