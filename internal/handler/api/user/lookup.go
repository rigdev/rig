package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user"
)

func (h Handler) GetByIdentifier(ctx context.Context, req *connect.Request[user.GetByIdentifierRequest]) (*connect.Response[user.GetByIdentifierResponse], error) {
	u, err := h.us.GetUserByIdentifier(ctx, req.Msg.GetIdentifier())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(
		&user.GetByIdentifierResponse{
			User: u,
		},
	), nil
}
