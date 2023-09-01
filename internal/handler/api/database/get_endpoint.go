package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) GetEndpoint(ctx context.Context, req *connect.Request[database.GetEndpointRequest]) (*connect.Response[database.GetEndpointResponse], error) {
	dbId, err := uuid.Parse(req.Msg.GetDatabaseId())
	if err != nil {
		return nil, err
	}

	endpoint, dbName, err := h.ds.GetDatabaseEndpoint(ctx, dbId, req.Msg.GetClientId(), req.Msg.GetClientSecret())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.GetEndpointResponse{Endpoint: endpoint, DatabaseName: dbName}), nil
}
