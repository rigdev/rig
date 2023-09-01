package storage

import (
	context "context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
)

func (h Handler) LookupProvider(ctx context.Context, req *connect.Request[storage.LookupProviderRequest]) (*connect.Response[storage.LookupProviderResponse], error) {
	providerID, provider, err := h.ss.LookupProvider(ctx, req.Msg.GetName())
	if err != nil {
		return nil, err
	}

	return &connect.Response[storage.LookupProviderResponse]{
		Msg: &storage.LookupProviderResponse{
			ProviderId: providerID.String(),
			Provider:   provider,
		},
	}, nil
}
