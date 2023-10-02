package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
)

func (h *Handler) ListInstances(ctx context.Context, req *connect.Request[capsule.ListInstancesRequest]) (*connect.Response[capsule.ListInstancesResponse], error) {
	it, total, err := h.cs.ListInstances(ctx, req.Msg.GetCapsuleId(), req.Msg.GetPagination())
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

func (h *Handler) ListInstanceStatuses(ctx context.Context, req *connect.Request[capsule.ListInstanceStatusesRequest]) (*connect.Response[capsule.ListInstanceStatusesResponse], error) {
	return nil, errors.UnimplementedErrorf("not implemented")
}
