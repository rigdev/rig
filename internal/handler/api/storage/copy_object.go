package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
)

func (h *Handler) CopyObject(ctx context.Context, req *connect.Request[storage.CopyObjectRequest]) (*connect.Response[storage.CopyObjectResponse], error) {
	err := h.ss.CopyObject(ctx, req.Msg.GetToBucket(), req.Msg.GetToPath(), req.Msg.GetFromBucket(), req.Msg.GetFromPath())
	if err != nil {
		return nil, err
	}
	return &connect.Response[storage.CopyObjectResponse]{
		Msg: &storage.CopyObjectResponse{},
	}, nil
}
