package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
)

// GetObject implements storageconnect.ServiceHandler
func (h *Handler) GetObject(ctx context.Context, req *connect.Request[storage.GetObjectRequest]) (*connect.Response[storage.GetObjectResponse], error) {
	object, err := h.ss.GetObject(ctx, req.Msg.Bucket, req.Msg.Path)
	if err != nil {
		return nil, err
	}
	return &connect.Response[storage.GetObjectResponse]{
		Msg: &storage.GetObjectResponse{
			Object: object,
		},
	}, nil
}
