package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
)

func (h Handler) UnlinkBucket(ctx context.Context, req *connect.Request[storage.UnlinkBucketRequest]) (*connect.Response[storage.UnlinkBucketResponse], error) {
	err := h.ss.UnlinkBucket(ctx, req.Msg.GetBucket())
	if err != nil {
		return nil, err
	}

	return &connect.Response[storage.UnlinkBucketResponse]{
		Msg: &storage.UnlinkBucketResponse{},
	}, nil
}
