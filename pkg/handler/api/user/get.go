package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/uuid"
)

// Get retries a user from the database from a specific project.
func (h *Handler) Get(ctx context.Context, req *connect.Request[user.GetRequest]) (*connect.Response[user.GetResponse], error) {
	uid, err := uuid.Parse(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}

	u, err := h.us.GetUser(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &connect.Response[user.GetResponse]{Msg: &user.GetResponse{User: u}}, nil
}
