package service_account

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) Delete(ctx context.Context, req *connect.Request[service_account.DeleteRequest]) (*connect.Response[service_account.DeleteResponse], error) {
	sid, err := uuid.Parse(req.Msg.GetServiceAccountId())
	if err != nil {
		return nil, err
	}

	if err := h.as.DeleteServiceAccount(ctx, sid); err != nil {
		return nil, err
	}

	return &connect.Response[service_account.DeleteResponse]{}, nil
}
