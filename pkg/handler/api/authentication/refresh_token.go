package authentication

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
)

// RefreshToken issues a new access/refresh token pair from the refresh token.
func (h *Handler) RefreshToken(ctx context.Context, req *connect.Request[authentication.RefreshTokenRequest]) (*connect.Response[authentication.RefreshTokenResponse], error) {
	t, err := h.as.RefreshToken(ctx, req.Msg.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &connect.Response[authentication.RefreshTokenResponse]{
		Msg: &authentication.RefreshTokenResponse{
			Token: t,
		},
	}, nil
}
