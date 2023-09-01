package storage

import (
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage/storageconnect"
	storage_service "github.com/rigdev/rig/internal/service/storage"
	"go.uber.org/fx"
)

type Handler struct {
	ss *storage_service.Service
}

type storageParams struct {
	fx.In
	Serv *storage_service.Service
}

func New(p storageParams) *Handler {
	return &Handler{
		ss: p.Serv,
	}
}

func (h *Handler) ServiceName() string {
	return storageconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return storageconnect.NewServiceHandler(h, opts...)
}
