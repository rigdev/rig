package capsule

import (
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule/capsuleconnect"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/internal/service/capsule"
	"github.com/rigdev/rig/internal/service/metrics"
)

type Handler struct {
	cs *capsule.Service
	ms metrics.Service
	pr repository.Project
}

// New returns an impementation of the proto Users.Users interface (service).
func New(cs *capsule.Service, pr repository.Project, ms metrics.Service) *Handler {
	return &Handler{
		cs: cs,
		ms: ms,
		pr: pr,
	}
}

func (h *Handler) ServiceName() string {
	return capsuleconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return capsuleconnect.NewServiceHandler(h, opts...)
}
