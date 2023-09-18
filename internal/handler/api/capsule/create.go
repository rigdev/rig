package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) Create(ctx context.Context, req *connect.Request[capsule.CreateRequest]) (*connect.Response[capsule.CreateResponse], error) {
	capsuleID, err := h.cs.CreateCapsule(ctx, req.Msg.GetName(), req.Msg.GetInitializers())
	if err != nil {
		return nil, err
	}

	return &connect.Response[capsule.CreateResponse]{
		Msg: &capsule.CreateResponse{
			CapsuleId: capsuleID,
		},
	}, nil
}
