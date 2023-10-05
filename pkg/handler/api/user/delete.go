package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/uuid"
)

// Delete will delete a user from the database in valid project.
func (h *Handler) Delete(ctx context.Context, req *connect.Request[user.DeleteRequest]) (*connect.Response[user.DeleteResponse], error) {
	uid, err := uuid.Parse(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}

	if err := h.us.DeleteUser(ctx, uid); err != nil {
		return nil, err
	}

	return &connect.Response[user.DeleteResponse]{}, nil
}
