package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) DeleteTable(ctx context.Context, req *connect.Request[database.DeleteTableRequest]) (*connect.Response[database.DeleteTableResponse], error) {
	dbId, err := uuid.Parse(req.Msg.GetDatabaseId())
	if err != nil {
		return nil, err
	}

	if err := h.ds.DeleteTable(ctx, dbId, req.Msg.GetTableName()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.DeleteTableResponse{}), nil
}
