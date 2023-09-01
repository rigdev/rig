package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) CreateTable(ctx context.Context, req *connect.Request[database.CreateTableRequest]) (*connect.Response[database.CreateTableResponse], error) {
	dbId, err := uuid.Parse(req.Msg.GetDatabaseId())
	if err != nil {
		return nil, err
	}

	if err := h.ds.CreateTable(ctx, dbId, req.Msg.GetTableName()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.CreateTableResponse{}), nil
}
