package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	api_database "github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) Delete(ctx context.Context, req *connect.Request[api_database.DeleteRequest]) (*connect.Response[api_database.DeleteResponse], error) {
	dbId, err := uuid.Parse(req.Msg.GetDatabaseId())
	if err != nil {
		return nil, err
	}

	if err := h.ds.Delete(ctx, dbId); err != nil {
		return nil, err
	}
	return connect.NewResponse(&api_database.DeleteResponse{}), nil
}
