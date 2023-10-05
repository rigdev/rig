package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/uuid"
)

// Get fetches the requested group from the database.
func (h *Handler) Get(ctx context.Context, req *connect.Request[group.GetRequest]) (resp *connect.Response[group.GetResponse], err error) {
	gid, err := uuid.Parse(req.Msg.GetGroupId())
	if err != nil {
		return nil, err
	}

	g, err := h.gs.Get(ctx, gid)
	if err != nil {
		return nil, err
	}
	return &connect.Response[group.GetResponse]{Msg: &group.GetResponse{Group: g}}, nil
}
