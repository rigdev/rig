package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/iterator"
)

func (h *Handler) ListBuilds(ctx context.Context, req *connect.Request[capsule.ListBuildsRequest]) (*connect.Response[capsule.ListBuildsResponse], error) {
	it, total, err := h.cs.ListBuilds(ctx, req.Msg.GetCapsuleId(), req.Msg.GetPagination())
	if err != nil {
		return nil, err
	}

	bs, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.ListBuildsResponse]{
		Msg: &capsule.ListBuildsResponse{
			Builds: bs,
			Total:  total,
		},
	}, nil
}
