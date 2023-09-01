package mongo

import (
	"context"

	"github.com/rigdev/rig/internal/repository/user/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
)

func (r *MongoRepository) DeleteOauth2Links(ctx context.Context, userID uuid.UUID) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	filter := schema.GetUserIDFilter(projectID, userID)
	_, err = r.Oauth2Col.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
