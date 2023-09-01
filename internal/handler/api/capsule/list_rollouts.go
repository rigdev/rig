package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h Handler) ListRollouts(ctx context.Context, req *connect.Request[capsule.ListRolloutsRequest]) (*connect.Response[capsule.ListRolloutsResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	it, total, err := h.cs.ListRollouts(ctx, cid, req.Msg.GetPagination())
	if err != nil {
		return nil, err
	}

	rollouts, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(
		&capsule.ListRolloutsResponse{
			Rollouts: rollouts,
			Total:    total,
		},
	), nil
}
