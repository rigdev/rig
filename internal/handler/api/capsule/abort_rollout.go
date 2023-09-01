package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) AbortRollout(ctx context.Context, req *connect.Request[capsule.AbortRolloutRequest]) (*connect.Response[capsule.AbortRolloutResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	if err := h.cs.AbortRollout(ctx, cid, req.Msg.GetRolloutId()); err != nil {
		return nil, err
	}

	return &connect.Response[capsule.AbortRolloutResponse]{}, nil
}
