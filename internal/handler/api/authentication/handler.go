package authentication

import (
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication/authenticationconnect"
	"github.com/rigdev/rig/internal/oauth2"
	"github.com/rigdev/rig/internal/service/auth"
	"github.com/rigdev/rig/internal/service/project"
	"github.com/rigdev/rig/internal/service/user"
	"go.uber.org/fx"
)

type Handler struct {
	ps     project.Service
	as     *auth.Service
	us     user.Service
	Oauth2 *oauth2.Providers
}

type sdkParams struct {
	fx.In
	ProjService project.Service
	AuthService *auth.Service
	UserService user.Service
	Oauth2      *oauth2.Providers
}

func New(p sdkParams) *Handler {
	return &Handler{
		ps:     p.ProjService,
		as:     p.AuthService,
		us:     p.UserService,
		Oauth2: p.Oauth2,
	}
}

func (h *Handler) ServiceName() string {
	return authenticationconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return authenticationconnect.NewServiceHandler(h, opts...)
}
