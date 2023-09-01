package storage_http

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	storage_service "github.com/rigdev/rig/internal/service/storage"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/service"
	"go.uber.org/fx"
)

type DownloadHandler struct {
	ss *storage_service.Service
}

type downloadParams struct {
	fx.In
	Serv *storage_service.Service
}

func NewDownloadHandler(p downloadParams) *DownloadHandler {
	return &DownloadHandler{
		ss: p.Serv,
	}
}

func (h *DownloadHandler) Build() (string, string, service.HandlerFunc) {
	return http.MethodGet, "/api/v1/storage/{bucket}/*", h.download
}

func (h *DownloadHandler) download(w http.ResponseWriter, r *http.Request) error {
	bucket := chi.URLParam(r, "bucket")
	if bucket == "" {
		return errors.InvalidArgumentErrorf("missing bucket name")
	}

	objectPath := chi.URLParam(r, "*")
	if objectPath == "" {
		return errors.InvalidArgumentErrorf("missing object path")
	}

	if strings.HasSuffix(objectPath, "/") {
		return errors.InvalidArgumentErrorf("object path cannot be a folder")
	}

	o, err := h.ss.GetObject(r.Context(), bucket, objectPath)
	if err != nil {
		return err
	}

	dr, err := h.ss.DownloadObject(r.Context(), bucket, objectPath)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Length", fmt.Sprint(o.Size))
	w.Header().Set("Content-Type", o.GetContentType())

	if _, err := io.Copy(w, dr); err != nil {
		return err
	}

	return nil
}
