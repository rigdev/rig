package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/iterator"
)

func (h Handler) ListProviders(ctx context.Context, req *connect.Request[storage.ListProvidersRequest]) (*connect.Response[storage.ListProvidersResponse], error) {
	it, total, err := h.ss.ListProviders(ctx, req.Msg.GetPagination())
	if err != nil {
		return nil, err
	}

	ps, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}

	return &connect.Response[storage.ListProvidersResponse]{
		Msg: &storage.ListProvidersResponse{
			Providers: ps,
			Total:     total,
		},
	}, nil
}
