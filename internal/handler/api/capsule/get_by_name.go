package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) GetByName(ctx context.Context, req *connect.Request[capsule.GetByNameRequest]) (*connect.Response[capsule.GetByNameResponse], error) {
	cap, err := h.cs.GetCapsuleByName(ctx, req.Msg.GetName())
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.GetByNameResponse]{
		Msg: &capsule.GetByNameResponse{
			Capsule: cap,
		},
	}, nil
}
