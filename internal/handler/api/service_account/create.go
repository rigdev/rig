package service_account

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
)

func (h *Handler) Create(ctx context.Context, req *connect.Request[service_account.CreateRequest]) (*connect.Response[service_account.CreateResponse], error) {
	sa, cID, cs, err := h.as.CreateServiceAccount(ctx, req.Msg.GetName(), false)
	if err != nil {
		return nil, err
	}

	return &connect.Response[service_account.CreateResponse]{
		Msg: &service_account.CreateResponse{
			ServiceAccount: sa,
			ClientId:       cID,
			ClientSecret:   cs,
		},
	}, nil
}
