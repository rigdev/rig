package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) Deploy(ctx context.Context, req *connect.Request[capsule.DeployRequest]) (*connect.Response[capsule.DeployResponse], error) {

	rolloutID, err := h.cs.Deploy(ctx, req.Msg.GetCapsuleId(), req.Msg.GetChanges())
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.DeployResponse]{
		Msg: &capsule.DeployResponse{
			RolloutId: rolloutID,
		},
	}, nil
}
