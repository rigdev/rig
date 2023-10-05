package service_account

import (
	"context"
	"io"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
)

func (h *Handler) List(ctx context.Context, req *connect.Request[service_account.ListRequest]) (*connect.Response[service_account.ListResponse], error) {
	it, err := h.as.ListServiceAccounts(ctx)
	if err != nil {
		return nil, err
	}

	res := &connect.Response[service_account.ListResponse]{
		Msg: &service_account.ListResponse{},
	}

	for {
		e, err := it.Next()
		if err == io.EOF {
			return res, nil
		} else if err != nil {
			return nil, err
		}

		res.Msg.ServiceAccounts = append(res.Msg.ServiceAccounts, e)
	}
}
