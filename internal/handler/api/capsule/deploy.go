package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) Deploy(ctx context.Context, req *connect.Request[capsule.DeployRequest]) (*connect.Response[capsule.DeployResponse], error) {
	if err := h.cs.Deploy(ctx, req.Msg.GetCapsuleId(), req.Msg.GetChanges()); err != nil {
		return nil, err
	}

	return &connect.Response[capsule.DeployResponse]{}, nil
}
