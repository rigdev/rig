package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h Handler) GetProvider(ctx context.Context, req *connect.Request[storage.GetProviderRequest]) (*connect.Response[storage.GetProviderResponse], error) {
	pid, err := uuid.Parse(req.Msg.GetProviderId())
	if err != nil {
		return nil, err
	}

	p, err := h.ss.GetProvider(ctx, pid)
	if err != nil {
		return nil, err
	}

	return &connect.Response[storage.GetProviderResponse]{
		Msg: &storage.GetProviderResponse{
			Provider: p,
		},
	}, nil
}
