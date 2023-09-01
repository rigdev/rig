package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) ListInstances(ctx context.Context, req *connect.Request[capsule.ListInstancesRequest]) (*connect.Response[capsule.ListInstancesResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	it, total, err := h.cs.ListInstances(ctx, cid, req.Msg.GetPagination())
	if err != nil {
		return nil, err
	}

	is, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.ListInstancesResponse]{
		Msg: &capsule.ListInstancesResponse{
			Instances: is,
			Total:     total,
		},
	}, nil
}
