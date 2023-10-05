package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/iterator"
)

// List returns a list of groups between "from" and "to".
func (h *Handler) List(ctx context.Context, req *connect.Request[group.ListRequest]) (*connect.Response[group.ListResponse], error) {
	it, total, err := h.gs.List(ctx, req.Msg.GetPagination(), req.Msg.GetSearch())
	if err != nil {
		return nil, err
	}

	gs, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}

	return &connect.Response[group.ListResponse]{
		Msg: &group.ListResponse{
			Groups: gs,
			Total:  total,
		},
	}, nil
}
