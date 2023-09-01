package settings

import (
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project/settings/settingsconnect"
	"github.com/rigdev/rig/internal/service/project"
	"go.uber.org/fx"
)

type Handler struct {
	ps project.Service
}

type newParams struct {
	fx.In

	ProjectService project.Service
}

// New returns an implementation of the proto Users.Users interface (service).
func New(p newParams) *Handler {
	return &Handler{
		ps: p.ProjectService,
	}
}

func (h *Handler) ServiceName() string {
	return settingsconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return settingsconnect.NewServiceHandler(h, opts...)
}
