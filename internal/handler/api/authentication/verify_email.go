package authentication

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) VerifyEmail(ctx context.Context, req *connect.Request[authentication.VerifyEmailRequest]) (*connect.Response[authentication.VerifyEmailResponse], error) {
	pID, err := uuid.Parse(req.Msg.GetProjectId())
	if err != nil {
		return nil, errors.InvalidArgumentErrorf("invalid project ID")
	}
	ctx = auth.WithProjectID(ctx, pID)
	if err := h.as.VerifyEmail(ctx, req.Msg.GetEmail(), req.Msg.GetCode()); err != nil {
		return nil, err
	}

	return &connect.Response[authentication.VerifyEmailResponse]{}, nil
}
