package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) CreateCredentials(ctx context.Context, req *connect.Request[database.CreateCredentialsRequest]) (*connect.Response[database.CreateCredentialsResponse], error) {
	dbId, err := uuid.Parse(req.Msg.GetDatabaseId())
	if err != nil {
		return nil, err
	}

	clientSecret, err := h.ds.CreateCredentials(ctx, req.Msg.GetClientId(), dbId)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.CreateCredentialsResponse{
		ClientSecret: clientSecret,
	}), nil
}
