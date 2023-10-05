package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) RestartInstance(ctx context.Context, req *connect.Request[capsule.RestartInstanceRequest]) (*connect.Response[capsule.RestartInstanceResponse], error) {
	if err := h.cs.RestartInstance(ctx, req.Msg.GetCapsuleId(), req.Msg.GetInstanceId()); err != nil {
		return nil, err
	}

	return &connect.Response[capsule.RestartInstanceResponse]{}, nil
}
