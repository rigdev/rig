package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
)

func (h *Handler) Create(ctx context.Context, req *connect.Request[database.CreateRequest]) (*connect.Response[database.CreateResponse], error) {
	db, err := h.ds.Create(ctx, req.Msg.GetName(), req.Msg.GetConfig(), req.Msg.GetLinkTables())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&database.CreateResponse{
		Database: db,
	}), nil
}
