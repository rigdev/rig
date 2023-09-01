package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/iterator"
)

// ListObjects implements storageconnect.ServiceHandler
func (h *Handler) ListObjects(ctx context.Context, req *connect.Request[storage.ListObjectsRequest]) (*connect.Response[storage.ListObjectsResponse], error) {
	token, it, err := h.ss.ListObjects(ctx, req.Msg.GetBucket(), req.Msg.GetToken(), req.Msg.GetPrefix(), req.Msg.GetStartPath(), req.Msg.GetEndPath(), req.Msg.GetRecursive(), req.Msg.GetLimit())
	if err != nil {
		return nil, err
	}
	res, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}
	return &connect.Response[storage.ListObjectsResponse]{
		Msg: &storage.ListObjectsResponse{
			Token:   token,
			Results: res,
		},
	}, nil
}
