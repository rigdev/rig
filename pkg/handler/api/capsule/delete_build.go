package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) DeleteBuild(ctx context.Context, req *connect.Request[capsule.DeleteBuildRequest]) (*connect.Response[capsule.DeleteBuildResponse], error) {
	if err := h.cs.DeleteBuild(ctx, req.Msg.GetCapsuleId(), req.Msg.GetBuildId()); err != nil {
		return nil, err
	}

	return &connect.Response[capsule.DeleteBuildResponse]{}, nil
}
