package handler

import (
	"github.com/rigdev/rig/internal/handler/api/authentication"
	"github.com/rigdev/rig/internal/handler/api/capsule"
	"github.com/rigdev/rig/internal/handler/api/database"
	"github.com/rigdev/rig/internal/handler/api/group"
	"github.com/rigdev/rig/internal/handler/api/project"
	project_settings "github.com/rigdev/rig/internal/handler/api/project/settings"
	"github.com/rigdev/rig/internal/handler/api/service_account"
	"github.com/rigdev/rig/internal/handler/api/status_http"
	"github.com/rigdev/rig/internal/handler/api/storage"
	"github.com/rigdev/rig/internal/handler/api/storage_http"
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
		asGRPCHandler(user.New),
		asGRPCHandler(group.New),
		asGRPCHandler(project.New),
		asGRPCHandler(capsule.New),
		asGRPCHandler(authentication.New),
		asGRPCHandler(service_account.New),
		asGRPCHandler(storage.New),
		asGRPCHandler(database.New),
		asGRPCHandler(user_settings.New),
		asGRPCHandler(project_settings.New),
		asHTTPHandler(http.New),
		asHTTPHandler(storage_http.NewUploadHandler),
		asHTTPHandler(storage_http.NewDownloadHandler),
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
