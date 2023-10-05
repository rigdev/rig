package group

import (
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group/groupconnect"
	"github.com/rigdev/rig/internal/service/group"
	"github.com/rigdev/rig/internal/service/user"
	"go.uber.org/fx"
)

type Handler struct {
	us user.Service
	gs *group.Service
}

type groupParams struct {
	fx.In
	UserServ  user.Service
	GroupServ *group.Service
}

func New(p groupParams) *Handler {
	return &Handler{
		us: p.UserServ,
		gs: p.GroupServ,
	}
}

func (h *Handler) ServiceName() string {
	return groupconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return groupconnect.NewServiceHandler(h, opts...)
}
