package group

import (
	"context"
	"io"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/uuid"
)

// GetMembers implements groupconnect.ServiceHandler
func (h *Handler) ListMembers(ctx context.Context, req *connect.Request[group.ListMembersRequest]) (*connect.Response[group.ListMembersResponse], error) {
	gid, err := uuid.Parse(req.Msg.GetGroupId())
	if err != nil {
		return nil, err
	}

	it, total, err := h.gs.ListMembers(ctx, gid, req.Msg.Pagination)
	if err != nil {
		return nil, err
	}

	defer it.Close()
	res := &group.ListMembersResponse{
		Total: total,
	}

	for {
		p, err := it.Next()
		if err == io.EOF {
			return &connect.Response[group.ListMembersResponse]{Msg: res}, nil
		} else if err != nil {
			return nil, err
		}

		res.Members = append(res.Members, p)
	}
}
