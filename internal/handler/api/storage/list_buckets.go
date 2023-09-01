package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/iterator"
)

// ListBuckets implements storageconnect.ServiceHandler
func (h *Handler) ListBuckets(ctx context.Context, req *connect.Request[storage.ListBucketsRequest]) (*connect.Response[storage.ListBucketsResponse], error) {
	it, err := h.ss.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}
	res, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}

	return &connect.Response[storage.ListBucketsResponse]{
		Msg: &storage.ListBucketsResponse{
			Buckets: res,
		},
	}, nil
}
