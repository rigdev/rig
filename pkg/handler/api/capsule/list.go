package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/iterator"
)

func (h *Handler) List(ctx context.Context, req *connect.Request[capsule.ListRequest]) (*connect.Response[capsule.ListResponse], error) {
	it, total, err := h.cs.ListCapsules(ctx, req.Msg.GetPagination())
	if err != nil {
		return nil, err
	}

	capsules, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.ListResponse]{
		Msg: &capsule.ListResponse{
			Capsules: capsules,
			Total:    uint64(total),
		},
	}, nil
}
