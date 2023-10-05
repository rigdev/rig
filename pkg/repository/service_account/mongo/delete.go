package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/repository/service_account/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (r *MongoRepository) Delete(ctx context.Context, serviceAccountID uuid.UUID) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	filter := schema.GetServiceAccountIDFilter(projectID, serviceAccountID)
	result, err := r.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.NotFoundErrorf("service account not found")
	}

	return nil
}
