package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) Delete(ctx context.Context, req *connect.Request[capsule.DeleteRequest]) (*connect.Response[capsule.DeleteResponse], error) {
	cid, err := uuid.Parse(req.Msg.GetCapsuleId())
	if err != nil {
		return nil, err
	}

	if err := h.cs.DeleteCapsule(ctx, cid); err != nil {
		return nil, err
	}

	return &connect.Response[capsule.DeleteResponse]{}, nil
}
