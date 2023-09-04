package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
)

func (h *Handler) GetByName(ctx context.Context, req *connect.Request[database.GetByNameRequest]) (*connect.Response[database.GetByNameResponse], error) {
	db, _, err := h.ds.GetByName(ctx, req.Msg.GetName())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(
		&database.GetByNameResponse{
			Database: db,
		},
	), nil
}
