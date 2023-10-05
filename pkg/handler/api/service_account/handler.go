package service_account

import (
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/service_account/service_accountconnect"
	"github.com/rigdev/rig/internal/service/auth"
)

type Handler struct {
	as *auth.Service
}

func New(as *auth.Service) *Handler {
	return &Handler{
		as: as,
	}
}

func (h *Handler) ServiceName() string {
	return service_accountconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return service_accountconnect.NewServiceHandler(h, opts...)
}
