package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
)

// DeleteBucket implements storageconnect.ServiceHandler
func (h *Handler) DeleteBucket(ctx context.Context, req *connect.Request[storage.DeleteBucketRequest]) (*connect.Response[storage.DeleteBucketResponse], error) {
	err := h.ss.DeleteBucket(ctx, req.Msg.Bucket)
	if err != nil {
		return nil, err
	}
	return &connect.Response[storage.DeleteBucketResponse]{
		Msg: &storage.DeleteBucketResponse{},
	}, nil
}
