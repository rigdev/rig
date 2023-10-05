package status_http

import (
	"net/http"

	"github.com/rigdev/rig/pkg/service"
)

type StatusHandler struct{}

func NewStatusHandler() *StatusHandler {
	return &StatusHandler{}
}

func (h *StatusHandler) Build() (string, string, service.HandlerFunc) {
	return http.MethodGet, "/api/v1/status", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("OK"))
		return nil
	}
}
