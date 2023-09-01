package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) Update(ctx context.Context, req *connect.Request[database.UpdateRequest]) (*connect.Response[database.UpdateResponse], error) {
	dbId, err := uuid.Parse(req.Msg.GetDatabaseId())
	if err != nil {
		return nil, err
	}

	db, err := h.ds.Update(ctx, dbId, req.Msg.GetUpdates())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.UpdateResponse{
		Database: db,
	}), nil
}
