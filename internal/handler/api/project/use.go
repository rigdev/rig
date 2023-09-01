package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/pkg/uuid"
)

// Use generates a project token for use with projects.
func (h *Handler) Use(ctx context.Context, req *connect.Request[project.UseRequest]) (resp *connect.Response[project.UseResponse], err error) {
	pid, err := uuid.Parse(req.Msg.GetProjectId())
	if err != nil {
		return nil, err
	}

	token, err := h.as.UseProject(ctx, pid)
	if err != nil {
		return nil, err
	}
	return &connect.Response[project.UseResponse]{Msg: &project.UseResponse{
		ProjectToken: token,
	}}, nil
}
