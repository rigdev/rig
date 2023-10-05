package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/uuid"
)

// Update the group information.
func (h *Handler) Update(ctx context.Context, req *connect.Request[group.UpdateRequest]) (resp *connect.Response[group.UpdateResponse], err error) {
	gid, err := uuid.Parse(req.Msg.GetGroupId())
	if err != nil {
		return nil, err
	}
	if err := h.gs.Update(ctx, gid, req.Msg.GetUpdates()); err != nil {
		return nil, err
	}
	return &connect.Response[group.UpdateResponse]{}, nil
}
