package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
)

func (h Handler) GetByName(ctx context.Context, req *connect.Request[group.GetByNameRequest]) (*connect.Response[group.GetByNameResponse], error) {
	g, err := h.gs.GetByName(ctx, req.Msg.GetName())
	if err != nil {
		return nil, err
	}
	return &connect.Response[group.GetByNameResponse]{
		Msg: &group.GetByNameResponse{
			Group: g,
		},
	}, nil
}
