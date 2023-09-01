package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/uuid"
)

// CreateBucket implements storageconnect.ServiceHandler
func (h *Handler) CreateBucket(ctx context.Context, req *connect.Request[storage.CreateBucketRequest]) (*connect.Response[storage.CreateBucketResponse], error) {
	pid, err := uuid.Parse(req.Msg.GetProviderId())
	if err != nil {
		return nil, err
	}

	err = h.ss.CreateBucket(ctx, req.Msg.GetBucket(), req.Msg.GetProviderBucket(), req.Msg.GetRegion(), pid)
	if err != nil {
		return nil, err
	}
	return &connect.Response[storage.CreateBucketResponse]{
		Msg: &storage.CreateBucketResponse{},
	}, nil
}
