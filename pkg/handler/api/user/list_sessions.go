package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

// ListSessions returns information about all tokens for a specific user.
func (h *Handler) ListSessions(ctx context.Context, req *connect.Request[user.ListSessionsRequest]) (*connect.Response[user.ListSessionsResponse], error) {
	uid, err := uuid.Parse(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}

	it, err := h.as.ListSessions(ctx, uid)
	if err != nil {
		return nil, err
	}

	ss, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}

	return &connect.Response[user.ListSessionsResponse]{
		Msg: &user.ListSessionsResponse{
			Sessions: ss,
		},
	}, nil
}
