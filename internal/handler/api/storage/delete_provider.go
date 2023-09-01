package storage

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h Handler) DeleteProvider(ctx context.Context, req *connect.Request[storage.DeleteProviderRequest]) (*connect.Response[storage.DeleteProviderResponse], error) {
	pid, err := uuid.Parse(req.Msg.GetProviderId())
	if err != nil {
		return nil, err
	}

	if err := h.ss.DeleteProvider(ctx, pid); err != nil {
		return nil, err
	}

	return &connect.Response[storage.DeleteProviderResponse]{
		Msg: &storage.DeleteProviderResponse{},
	}, nil
}
