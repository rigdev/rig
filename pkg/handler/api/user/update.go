package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/uuid"
)

// Update updates profile and/or extra data information of a specific user; email, phone, and/or username.
func (h *Handler) Update(ctx context.Context, req *connect.Request[user.UpdateRequest]) (*connect.Response[user.UpdateResponse], error) {
	uid, err := uuid.Parse(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}

	if err := h.us.UpdateUser(ctx, uid, req.Msg.GetUpdates()); err != nil {
		return nil, err
	}

	return &connect.Response[user.UpdateResponse]{}, nil
}
