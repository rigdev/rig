package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) Get(ctx context.Context, req *connect.Request[capsule.GetRequest]) (*connect.Response[capsule.GetResponse], error) {
	c, err := h.cs.GetCapsule(ctx, req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.GetResponse]{
		Msg: &capsule.GetResponse{
			Capsule: c,
		},
	}, nil
}
