package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
)

// Delete deletes an entire project by dropping all collections in Users.
func (h *Handler) Delete(ctx context.Context, req *connect.Request[project.DeleteRequest]) (resp *connect.Response[project.DeleteResponse], err error) {
	if err := h.ps.DeleteProject(ctx); err != nil {
		return nil, err
	}

	return &connect.Response[project.DeleteResponse]{}, nil
}
