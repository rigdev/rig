package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (h *Handler) DeleteCredential(ctx context.Context, req *connect.Request[database.DeleteCredentialRequest]) (*connect.Response[database.DeleteCredentialResponse], error) {
	dbId, err := uuid.Parse(req.Msg.GetDatabaseId())
	if err != nil {
		return nil, err
	}

	err = h.ds.DeleteCredential(ctx, req.Msg.GetCredentialName(), dbId)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&database.DeleteCredentialResponse{}), nil
}
