package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) CreateCredential(ctx context.Context, req *connect.Request[database.CreateCredentialRequest]) (*connect.Response[database.CreateCredentialResponse], error) {
	dbId, err := uuid.Parse(req.Msg.GetDatabaseId())
	if err != nil {
		return nil, err
	}

	clientId, clientSecret, err := h.ds.CreateCredential(ctx, req.Msg.GetName(), dbId)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.CreateCredentialResponse{
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}), nil
}
