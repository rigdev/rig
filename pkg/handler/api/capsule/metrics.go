package capsule

import (
	"context"
	"io"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) CapsuleMetrics(ctx context.Context, req *connect.Request[capsule.CapsuleMetricsRequest]) (*connect.Response[capsule.CapsuleMetricsResponse], error) {
	it, err := h.ms.ListWhereCapsuleID(ctx, req.Msg.GetPagination(), req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	var ims []*capsule.InstanceMetrics

	for {
		m, err := it.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		ims = append(ims, m)
	}

	return &connect.Response[capsule.CapsuleMetricsResponse]{
		Msg: &capsule.CapsuleMetricsResponse{InstanceMetrics: ims},
	}, nil
}
