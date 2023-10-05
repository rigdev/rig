package project

import (
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project/projectconnect"
	"github.com/rigdev/rig/internal/service/auth"
	"github.com/rigdev/rig/internal/service/project"
	"go.uber.org/fx"
)

type Handler struct {
	ps project.Service
	as *auth.Service
}

type projectParams struct {
	fx.In
	ProjectService project.Service
	AuthService    *auth.Service
}

func New(p projectParams) *Handler {
	return &Handler{
		ps: p.ProjectService,
		as: p.AuthService,
	}
}

func (h *Handler) ServiceName() string {
	return projectconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return projectconnect.NewServiceHandler(h, opts...)
}
