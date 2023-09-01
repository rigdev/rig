package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
)

func (h Handler) CreateProvider(ctx context.Context, req *connect.Request[storage.CreateProviderRequest]) (*connect.Response[storage.CreateProviderResponse], error) {
	_, p, err := h.ss.CreateProvider(ctx, req.Msg.GetName(), req.Msg.GetConfig(), req.Msg.GetLinkBuckets())
	if err != nil {
		return nil, err
	}

	return &connect.Response[storage.CreateProviderResponse]{
		Msg: &storage.CreateProviderResponse{
			Provider: p,
		},
	}, nil
}
