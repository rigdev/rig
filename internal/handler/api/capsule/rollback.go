package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) Rollback(ctx context.Context, req *connect.Request[capsule.RollbackRequest]) (*connect.Response[capsule.RollbackResponse], error) {
	rollout, err := h.cs.Rollback(ctx, req.Msg.GetCapsuleId(), req.Msg.GetRolloutId())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&capsule.RollbackResponse{
		RolloutId: rollout,
	}), nil
}
