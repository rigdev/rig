package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/iterator"
)

func (h Handler) ListEvents(ctx context.Context, req *connect.Request[capsule.ListEventsRequest]) (*connect.Response[capsule.ListEventsResponse], error) {
	it, total, err := h.cs.ListEvents(ctx, req.Msg.GetCapsuleId(), req.Msg.GetRolloutId(), req.Msg.GetPagination())
	if err != nil {
		return nil, err
	}

	events, err := iterator.Collect(it)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(
		&capsule.ListEventsResponse{
			Events: events,
			Total:  total,
		},
	), nil
}
