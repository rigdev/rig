package allplugins

import (
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/plugins/builtin/annotations"
	argorollout "github.com/rigdev/rig/plugins/builtin/argo_rollout"
	"github.com/rigdev/rig/plugins/builtin/datadog"
	envmapping "github.com/rigdev/rig/plugins/builtin/env_mapping"
	envvarcsi "github.com/rigdev/rig/plugins/builtin/envvar_csi"
	googlesqlproxy "github.com/rigdev/rig/plugins/builtin/google_cloud_sql_auth_proxy"
	initcontainer "github.com/rigdev/rig/plugins/builtin/init_container"
	objectcreate "github.com/rigdev/rig/plugins/builtin/object_create"
	objecttemplate "github.com/rigdev/rig/plugins/builtin/object_template"
	"github.com/rigdev/rig/plugins/builtin/placement"
	"github.com/rigdev/rig/plugins/builtin/sidecar"
	"github.com/rigdev/rig/plugins/capsulesteps/cron_jobs"
	"github.com/rigdev/rig/plugins/capsulesteps/deployment"
	ingressroutes "github.com/rigdev/rig/plugins/capsulesteps/ingress_routes"
	"github.com/rigdev/rig/plugins/capsulesteps/service_account"
	"github.com/rigdev/rig/plugins/capsulesteps/service_monitor"
	"github.com/rigdev/rig/plugins/capsulesteps/vpa"
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
	objectcreate.Name:    &objectcreate.Plugin{},
	envvarcsi.Name:       &envvarcsi.Plugin{},
	argorollout.Name:     &argorollout.Plugin{},
}
