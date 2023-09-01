package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
)

// GetBucket implements storageconnect.ServiceHandler
func (h *Handler) GetBucket(ctx context.Context, req *connect.Request[storage.GetBucketRequest]) (*connect.Response[storage.GetBucketResponse], error) {
	bucket, err := h.ss.GetBucket(ctx, req.Msg.Bucket)
	if err != nil {
		return nil, err
	}
	return &connect.Response[storage.GetBucketResponse]{
		Msg: &storage.GetBucketResponse{
			Bucket: bucket,
		},
	}, nil
}
