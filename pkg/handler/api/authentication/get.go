package authentication

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
)

// Get fetches a user from the namespace using its id in the access token.
func (h *Handler) Get(ctx context.Context, req *connect.Request[authentication.GetRequest]) (*connect.Response[authentication.GetResponse], error) {
	c, err := auth.GetClaims(ctx)
	if err != nil {
		return nil, err
	}

	if c.GetSubjectType() != auth.SubjectTypeUser {
		return nil, errors.InvalidArgumentErrorf("not a user access token")
	}

	ctx = auth.WithProjectID(ctx, c.GetProjectID())
	u, err := h.us.GetUser(ctx, c.GetSubject())
	if err != nil {
		return nil, err
	}

	return &connect.Response[authentication.GetResponse]{
		Msg: &authentication.GetResponse{
			UserInfo: u.GetUserInfo(),
			UserId:   c.GetSubject().String(),
		},
	}, nil
}
