package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) CreateBuild(ctx context.Context, req *connect.Request[capsule.CreateBuildRequest]) (*connect.Response[capsule.CreateBuildResponse], error) {
	resp, err := h.cs.CreateBuild(
		ctx,
		req.Msg.GetCapsuleId(),
		req.Msg.GetImage(),
		req.Msg.GetDigest(),
		req.Msg.GetOrigin(),
		req.Msg.GetLabels(),
		!req.Msg.GetSkipImageCheck(),
	)
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.CreateBuildResponse]{
		Msg: &capsule.CreateBuildResponse{
			BuildId:         resp.BuildID,
			CreatedNewBuild: resp.CreatedNewBuild,
		},
	}, nil
}
