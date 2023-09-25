package handler

import (
	"github.com/rigdev/rig/internal/handler/api/authentication"
	"github.com/rigdev/rig/internal/handler/api/capsule"
	"github.com/rigdev/rig/internal/handler/api/cluster"
	"github.com/rigdev/rig/internal/handler/api/group"
	"github.com/rigdev/rig/internal/handler/api/project"
	project_settings "github.com/rigdev/rig/internal/handler/api/project/settings"
	"github.com/rigdev/rig/internal/handler/api/service_account"
	"github.com/rigdev/rig/internal/handler/api/status_http"
	"github.com/rigdev/rig/internal/handler/api/user"
	user_settings "github.com/rigdev/rig/internal/handler/api/user/settings"
	"github.com/rigdev/rig/internal/handler/http"
	"github.com/rigdev/rig/internal/handler/registry"
	"github.com/rigdev/rig/pkg/service"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"handler",
	fx.Provide(
		asGRPCHandler(authentication.New),
		asGRPCHandler(capsule.New),
		asGRPCHandler(cluster.New),
		asGRPCHandler(group.New),
		asGRPCHandler(project.New),
		asGRPCHandler(project_settings.New),
		asGRPCHandler(service_account.New),
		asGRPCHandler(user.New),
		asGRPCHandler(user_settings.New),
		asHTTPHandler(http.New),
		asHTTPHandler(status_http.NewStatusHandler),
		registry.NewServer,
	),
)

func asGRPCHandler(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(service.GRPCHandler)),
		fx.ResultTags(`group:"grpc_handlers"`),
	)
}

func asHTTPHandler(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(service.HTTPHandler)),
		fx.ResultTags(`group:"http_handlers"`),
	)
}
