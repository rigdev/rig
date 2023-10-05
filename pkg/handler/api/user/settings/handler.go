package settings

import (
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user/settings/settingsconnect"
	"github.com/rigdev/rig/internal/service/user"
	"go.uber.org/fx"
)

type Handler struct {
	us user.Service
}

type newParams struct {
	fx.In

	UserService user.Service
}

// New returns an implementation of the proto Users.Users interface (service).
func New(p newParams) *Handler {
	return &Handler{
		us: p.UserService,
	}
}

func (h *Handler) ServiceName() string {
	return settingsconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return settingsconnect.NewServiceHandler(h, opts...)
}
