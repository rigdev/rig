package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/pkg/auth"
)

// Update updates the profile of the project. This can be values such as the name and image of the project.
func (h *Handler) Update(ctx context.Context, req *connect.Request[project.UpdateRequest]) (resp *connect.Response[project.UpdateResponse], err error) {
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.ps.UpdateProject(ctx, projectId, req.Msg.GetUpdates()); err != nil {
		return nil, err
	}

	return &connect.Response[project.UpdateResponse]{}, nil
}
