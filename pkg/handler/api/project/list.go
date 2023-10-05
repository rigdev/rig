package project

import (
	"context"
	"io"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
)

// List implements projectconnect.ServiceHandler
func (h *Handler) List(ctx context.Context, req *connect.Request[project.ListRequest]) (*connect.Response[project.ListResponse], error) {
	it, count, err := h.ps.List(ctx, req.Msg.Pagination)
	if err != nil {
		return nil, err
	}

	defer it.Close()
	res := &project.ListResponse{
		Total: count,
	}

	for {
		p, err := it.Next()
		if err == io.EOF {
			return &connect.Response[project.ListResponse]{Msg: res}, nil
		} else if err != nil {
			return nil, err
		}

		res.Projects = append(res.Projects, p)
	}
}
