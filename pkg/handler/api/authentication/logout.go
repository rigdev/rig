package authentication

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
)

// Logout validates the access token and blocks it in the database for future attempts.
func (h *Handler) Logout(ctx context.Context, req *connect.Request[authentication.LogoutRequest]) (*connect.Response[authentication.LogoutResponse], error) {
	if err := h.as.Logout(ctx); err != nil {
		return nil, err
	}

	return &connect.Response[authentication.LogoutResponse]{}, nil
}
