package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) Update(ctx context.Context, req *connect.Request[capsule.UpdateRequest]) (*connect.Response[capsule.UpdateResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	if err := h.cs.UpdateCapsule(ctx, cid, req.Msg.Updates); err != nil {
		return nil, err
	}
	return &connect.Response[capsule.UpdateResponse]{}, nil
}
