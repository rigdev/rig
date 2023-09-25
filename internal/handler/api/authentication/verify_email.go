package authentication

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/pkg/auth"
)

func (h *Handler) VerifyEmail(ctx context.Context, req *connect.Request[authentication.VerifyEmailRequest]) (*connect.Response[authentication.VerifyEmailResponse], error) {
	pID := req.Msg.GetProjectId()
	ctx = auth.WithProjectID(ctx, pID)
	if err := h.as.VerifyEmail(ctx, req.Msg.GetEmail(), req.Msg.GetCode()); err != nil {
		return nil, err
	}

	return &connect.Response[authentication.VerifyEmailResponse]{}, nil
}
