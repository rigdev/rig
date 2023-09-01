package user

import (
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user/userconnect"
	"github.com/rigdev/rig/internal/service/auth"
	"github.com/rigdev/rig/internal/service/user"
	"go.uber.org/fx"
)

type Handler struct {
	us user.Service
	as *auth.Service
}

type newParams struct {
	fx.In

	UserService user.Service
	AuthService *auth.Service
}

// New returns an implementation of the proto Users.Users interface (service).
func New(p newParams) *Handler {
	return &Handler{
		us: p.UserService,
		as: p.AuthService,
	}
}

func (h *Handler) ServiceName() string {
	return userconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return userconnect.NewServiceHandler(h, opts...)
}
