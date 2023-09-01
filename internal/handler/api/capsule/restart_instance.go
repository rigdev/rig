package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) RestartInstance(ctx context.Context, req *connect.Request[capsule.RestartInstanceRequest]) (*connect.Response[capsule.RestartInstanceResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	if err := h.cs.RestartInstance(ctx, cid, req.Msg.GetInstanceId()); err != nil {
		return nil, err
	}

	return &connect.Response[capsule.RestartInstanceResponse]{}, nil
}
