package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/iterator"
)

// List fetches a list of users in the database together with a count of total users in the database.
func (h *Handler) List(ctx context.Context, req *connect.Request[user.ListRequest]) (*connect.Response[user.ListResponse], error) {
	it, total, err := h.us.List(ctx, req.Msg.GetPagination(), req.Msg.GetSearch())
	if err != nil {
		return nil, err
	}

	us, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}

	return &connect.Response[user.ListResponse]{
		Msg: &user.ListResponse{
			Users: us,
			Total: total,
		},
	}, nil
}
