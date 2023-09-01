package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) GetRollout(ctx context.Context, req *connect.Request[capsule.GetRolloutRequest]) (*connect.Response[capsule.GetRolloutResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	r, err := h.cs.GetRollout(ctx, cid, req.Msg.GetRolloutId())
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.GetRolloutResponse]{
		Msg: &capsule.GetRolloutResponse{
			Rollout: r,
		},
	}, nil
}
