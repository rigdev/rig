package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) CreateBuild(ctx context.Context, req *connect.Request[capsule.CreateBuildRequest]) (*connect.Response[capsule.CreateBuildResponse], error) {
	capsuleID, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	buildID, err := h.cs.CreateBuild(
		ctx,
		capsuleID,
		req.Msg.GetImage(),
		req.Msg.GetDigest(),
		req.Msg.GetOrigin(),
		req.Msg.GetLabels(),
	)
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.CreateBuildResponse]{
		Msg: &capsule.CreateBuildResponse{
			BuildId: buildID,
		},
	}, nil
}
