package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
)

// Create inserts a new group in the database.
func (h *Handler) Create(ctx context.Context, req *connect.Request[group.CreateRequest]) (resp *connect.Response[group.CreateResponse], err error) {
	g, err := h.gs.Create(ctx, req.Msg.GetInitializers())
	if err != nil {
		return nil, err
	}

	return &connect.Response[group.CreateResponse]{
		Msg: &group.CreateResponse{
			Group: g,
		},
	}, nil
}
