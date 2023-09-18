package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) AbortRollout(ctx context.Context, req *connect.Request[capsule.AbortRolloutRequest]) (*connect.Response[capsule.AbortRolloutResponse], error) {
	if err := h.cs.AbortRollout(ctx, req.Msg.GetCapsuleId(), req.Msg.GetRolloutId()); err != nil {
		return nil, err
	}

	return &connect.Response[capsule.AbortRolloutResponse]{}, nil
}
