package group

import (
	"context"
	"io"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/uuid"
)

// GetGroups implements groupconnect.ServiceHandler
func (h *Handler) ListGroupsForUser(ctx context.Context, req *connect.Request[group.ListGroupsForUserRequest]) (*connect.Response[group.ListGroupsForUserResponse], error) {
	uid, err := uuid.Parse(req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}

	it, count, err := h.gs.ListGroupsForUser(ctx, uid, req.Msg.GetPagination())
	if err != nil {
		return nil, err
	}

	defer it.Close()
	res := &group.ListGroupsForUserResponse{
		Total: count,
	}

	for {
		p, err := it.Next()
		if err == io.EOF {
			return &connect.Response[group.ListGroupsForUserResponse]{Msg: res}, nil
		} else if err != nil {
			return nil, err
		}

		res.Groups = append(res.Groups, p)
	}
}
