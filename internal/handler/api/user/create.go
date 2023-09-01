package user

import (
	"context"

	"github.com/bufbuild/connect-go"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
)

// Create inserts a user in the database using the specified project project.
func (h *Handler) Create(ctx context.Context, req *connect.Request[user.CreateRequest]) (*connect.Response[user.CreateResponse], error) {
	u, err := h.us.CreateUser(ctx, &model.RegisterMethod{Method: &model.RegisterMethod_System_{}}, req.Msg.GetInitializers())
	if err != nil {
		return nil, err
	}

	return &connect.Response[user.CreateResponse]{
		Msg: &user.CreateResponse{
			User: u,
		},
	}, nil
}
