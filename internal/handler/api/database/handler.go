package database

import (
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database/databaseconnect"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/internal/service/database"
)

type Handler struct {
	ds *database.Service
	pr repository.Project
}

func New(ds *database.Service, pr repository.Project) *Handler {
	return &Handler{
		ds: ds,
		pr: pr,
	}
}

func (h *Handler) ServiceName() string {
	return databaseconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return databaseconnect.NewServiceHandler(h, opts...)
}
