package authentication

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/pkg/auth"
)

// ResetPassword updates the users password if the provided code can be validated by the hash.
func (h *Handler) ResetPassword(ctx context.Context, req *connect.Request[authentication.ResetPasswordRequest]) (*connect.Response[authentication.ResetPasswordResponse], error) {
	pID := req.Msg.GetProjectId()
	ctx = auth.WithProjectID(ctx, pID)
	if err := h.as.ResetPassword(ctx, req.Msg.GetIdentifier(), req.Msg.GetCode(), req.Msg.GetNewPassword()); err != nil {
		return nil, err
	}
	return &connect.Response[authentication.ResetPasswordResponse]{}, nil
}
