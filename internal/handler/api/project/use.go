package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
)

// Use generates a project token for use with projects.
func (h *Handler) Use(ctx context.Context, req *connect.Request[project.UseRequest]) (resp *connect.Response[project.UseResponse], err error) {
	token, err := h.as.UseProject(ctx, req.Msg.GetProjectId())
	if err != nil {
		return nil, err
	}
	return &connect.Response[project.UseResponse]{Msg: &project.UseResponse{
		ProjectToken: token,
	}}, nil
}
