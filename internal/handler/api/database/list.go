package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/iterator"
)

func (h *Handler) List(ctx context.Context, req *connect.Request[database.ListRequest]) (*connect.Response[database.ListResponse], error) {
	it, total, err := h.ds.List(ctx, req.Msg.GetPagination())
	if err != nil {
		return nil, err
	}
	ds, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.ListResponse{
		Databases: ds,
		Total:     total,
	}), nil
}
