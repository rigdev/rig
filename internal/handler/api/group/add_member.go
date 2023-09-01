package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

// AddUser implements groupconnect.ServiceHandler
func (h *Handler) AddMember(ctx context.Context, req *connect.Request[group.AddMemberRequest]) (*connect.Response[group.AddMemberResponse], error) {
	gid, err := uuid.Parse(req.Msg.GetGroupId())
	if err != nil {
		return nil, err
	}
	if len(req.Msg.GetUserIds()) == 0 {
		return nil, errors.InvalidArgumentErrorf("missing user ids")
	}
	var ids []uuid.UUID
	for _, id := range req.Msg.GetUserIds() {
		gid, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, gid)
	}

	if err := h.gs.AddMembers(ctx, gid, ids); err != nil {
		return nil, err
	}
	return &connect.Response[group.AddMemberResponse]{}, nil
}
