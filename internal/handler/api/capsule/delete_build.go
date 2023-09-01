package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) DeleteBuild(ctx context.Context, req *connect.Request[capsule.DeleteBuildRequest]) (*connect.Response[capsule.DeleteBuildResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	if err := h.cs.DeleteBuild(ctx, cid, req.Msg.GetBuildId()); err != nil {
		return nil, err
	}

	return &connect.Response[capsule.DeleteBuildResponse]{}, nil
}
