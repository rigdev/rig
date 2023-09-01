package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
)

// DeleteObject implements storageconnect.ServiceHandler
func (h *Handler) DeleteObject(ctx context.Context, req *connect.Request[storage.DeleteObjectRequest]) (*connect.Response[storage.DeleteObjectResponse], error) {
	err := h.ss.DeleteObject(ctx, req.Msg.Bucket, req.Msg.Path)
	if err != nil {
		return nil, err
	}
	return &connect.Response[storage.DeleteObjectResponse]{
		Msg: &storage.DeleteObjectResponse{},
	}, nil
}
