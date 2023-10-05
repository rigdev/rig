package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
)

// GetConfig retries a config from the database from a specific project and also returns available providers.
func (h *Handler) Get(ctx context.Context, req *connect.Request[project.GetRequest]) (resp *connect.Response[project.GetResponse], err error) {
	p, err := h.ps.GetProject(ctx)
	if err != nil {
		return nil, err
	}
	return &connect.Response[project.GetResponse]{Msg: &project.GetResponse{
		Project: p,
	}}, nil
}
