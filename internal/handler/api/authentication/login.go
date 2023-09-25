package authentication

import (
	"context"
	"reflect"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/telemetry"
)

// Login validates a user credentials and issues a access/refresh token pair used to validate requests.
func (h *Handler) Login(ctx context.Context, req *connect.Request[authentication.LoginRequest]) (*connect.Response[authentication.LoginResponse], error) {
	switch v := req.Msg.GetMethod().(type) {
	case *authentication.LoginRequest_UserPassword:
		pID := v.UserPassword.GetProjectId()
		ctx = auth.WithProjectID(ctx, pID)
		userID, u, t, err := h.as.LoginUserPassword(ctx, v.UserPassword.GetIdentifier(), v.UserPassword.GetPassword())
		if err != nil {
			return nil, err
		}

		if pID == auth.RigProjectID {
			if u.GetEmail() != "" {
				telemetry.SetUserEmail(ctx, userID, u.GetEmail())
			}
			if u.GetUsername() != "" {
				telemetry.SetUserUsername(ctx, userID, u.GetUsername())
			}
		}

		return &connect.Response[authentication.LoginResponse]{
			Msg: &authentication.LoginResponse{
				Token:    t,
				UserId:   userID.String(),
				UserInfo: u,
			},
		}, nil
	case *authentication.LoginRequest_ClientCredentials:
		t, err := h.as.LoginClientCredentials(ctx, v.ClientCredentials.GetClientId(), v.ClientCredentials.GetClientSecret())
		if err != nil {
			return nil, err
		}

		return &connect.Response[authentication.LoginResponse]{
			Msg: &authentication.LoginResponse{
				Token: t,
			},
		}, nil
	default:
		return nil, errors.InvalidArgumentErrorf("invalid login method '%v'", reflect.TypeOf(v))
	}
}
