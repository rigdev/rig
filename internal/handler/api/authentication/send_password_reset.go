package authentication

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

// SendResetPasswordEmail sends an email to the user with a code used to reset his/her password.
func (h *Handler) SendPasswordReset(ctx context.Context, req *connect.Request[authentication.SendPasswordResetRequest]) (*connect.Response[authentication.SendPasswordResetResponse], error) {
	pID, err := uuid.Parse(req.Msg.GetProjectId())
	if err != nil {
		return nil, errors.InvalidArgumentErrorf("invalid project ID")
	}
	ctx = auth.WithProjectID(ctx, pID)
	if err := h.as.SendPasswordReset(ctx, req.Msg.GetIdentifier()); err != nil {
		return nil, err
	}
	return &connect.Response[authentication.SendPasswordResetResponse]{}, nil
}
