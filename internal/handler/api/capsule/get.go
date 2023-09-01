package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) Get(ctx context.Context, req *connect.Request[capsule.GetRequest]) (*connect.Response[capsule.GetResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	c, err := h.cs.GetCapsule(ctx, cid)
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.GetResponse]{
		Msg: &capsule.GetResponse{
			Capsule: c,
		},
	}, nil
}
