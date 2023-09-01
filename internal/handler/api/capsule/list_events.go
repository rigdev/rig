package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h Handler) ListEvents(ctx context.Context, req *connect.Request[capsule.ListEventsRequest]) (*connect.Response[capsule.ListEventsResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	it, total, err := h.cs.ListEvents(ctx, cid, req.Msg.GetRolloutId(), req.Msg.GetPagination())
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
