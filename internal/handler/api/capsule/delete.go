package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

func (h *Handler) Delete(ctx context.Context, req *connect.Request[capsule.DeleteRequest]) (*connect.Response[capsule.DeleteResponse], error) {

	if err := h.cs.DeleteCapsule(ctx, req.Msg.GetCapsuleId()); err != nil {
		return nil, err
	}

	return &connect.Response[capsule.DeleteResponse]{}, nil
}
