package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
)

func (h *Handler) Create(ctx context.Context, req *connect.Request[database.CreateRequest]) (*connect.Response[database.CreateResponse], error) {
	databaseID, db, err := h.ds.Create(ctx, req.Msg.GetType(), req.Msg.GetInitializers())
	if err != nil {
		return nil, err
	}
	clientId, clientSecret, err := h.ds.CreateCredential(ctx, "Default Credential", databaseID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.CreateResponse{
		Database:     db,
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}), nil
}
