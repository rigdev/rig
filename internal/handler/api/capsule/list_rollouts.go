package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/iterator"
)

func (h Handler) ListRollouts(ctx context.Context, req *connect.Request[capsule.ListRolloutsRequest]) (*connect.Response[capsule.ListRolloutsResponse], error) {
	it, total, err := h.cs.ListRollouts(ctx, req.Msg.GetCapsuleId(), req.Msg.GetPagination())
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
