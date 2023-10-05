package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/uuid"
)

// RemoveUser implements groupconnect.ServiceHandler
func (h *Handler) RemoveMember(ctx context.Context, req *connect.Request[group.RemoveMemberRequest]) (*connect.Response[group.RemoveMemberResponse], error) {
	gid, err := uuid.Parse(req.Msg.GetGroupId())
	if err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(req.Msg.GetGroupId())
	if err != nil {
		return nil, err
	}

	if err := h.gs.RemoveMember(ctx, gid, uid); err != nil {
		return nil, err
	}
	return &connect.Response[group.RemoveMemberResponse]{}, nil
}
