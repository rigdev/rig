package storage_http

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	storage_service "github.com/rigdev/rig/internal/service/storage"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/service"
	"go.uber.org/fx"
)

type UploadHandler struct {
	ss *storage_service.Service
}

type uploadParams struct {
	fx.In
	Serv *storage_service.Service
}

func NewUploadHandler(p uploadParams) *UploadHandler {
	return &UploadHandler{
		ss: p.Serv,
	}
}

func (h *UploadHandler) Build() (string, string, service.HandlerFunc) {
	return http.MethodPut, "/api/v1/storage/{bucket}/*", h.upload
}

func (h *UploadHandler) upload(w http.ResponseWriter, r *http.Request) error {
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

	if r.ContentLength < 0 {
		return errors.InvalidArgumentErrorf("missing Content-Length header")
	}

	_, _, err := h.ss.UploadObject(r.Context(), r.Body, &storage.UploadObjectRequest_Metadata{
		Bucket:      bucket,
		Path:        objectPath,
		Size:        uint64(r.ContentLength),
		ContentType: r.Header.Get("Content-Type"),
	})
	return err
}
