package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
)

// Create creates and sets up a project
func (h *Handler) Create(ctx context.Context, req *connect.Request[project.CreateRequest]) (resp *connect.Response[project.CreateResponse], err error) {
	p, err := h.ps.CreateProject(ctx, req.Msg.GetInitializers())
	if err != nil {
		return nil, err
	}

	return &connect.Response[project.CreateResponse]{
		Msg: &project.CreateResponse{
			Project: p,
		},
	}, nil
}
