package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) Deploy(ctx context.Context, req *connect.Request[capsule.DeployRequest]) (*connect.Response[capsule.DeployResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	if err := h.cs.Deploy(ctx, cid, req.Msg.GetChanges()); err != nil {
		return nil, err
	}

	return &connect.Response[capsule.DeployResponse]{}, nil
}
