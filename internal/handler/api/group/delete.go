package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/uuid"
)

// Delete removes the group from the database as well in all users.
func (h *Handler) Delete(ctx context.Context, req *connect.Request[group.DeleteRequest]) (resp *connect.Response[group.DeleteResponse], err error) {
	gid, err := uuid.Parse(req.Msg.GetGroupId())
	if err != nil {
		return nil, err
	}

	if err := h.gs.Delete(ctx, gid); err != nil {
		return nil, err
	}

	return &connect.Response[group.DeleteResponse]{}, nil
}
