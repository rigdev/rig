package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) ListTables(ctx context.Context, req *connect.Request[database.ListTablesRequest]) (*connect.Response[database.ListTablesResponse], error) {
	dbId, err := uuid.Parse(req.Msg.GetDatabaseId())
	if err != nil {
		return nil, err
	}

	tables, err := h.ds.ListTables(ctx, dbId)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.ListTablesResponse{
		Tables: tables,
		Total:  uint64(len(tables)),
	}), nil
}
