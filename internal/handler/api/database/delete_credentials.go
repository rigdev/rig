package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) DeleteCredentials(ctx context.Context, req *connect.Request[database.DeleteCredentialsRequest]) (*connect.Response[database.DeleteCredentialsResponse], error) {
	dbId, err := uuid.Parse(req.Msg.GetDatabaseId())
	if err != nil {
		return nil, err
	}

	err = h.ds.DeleteCredentials(ctx, req.Msg.GetName(), dbId)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.DeleteCredentialsResponse{}), nil
}
