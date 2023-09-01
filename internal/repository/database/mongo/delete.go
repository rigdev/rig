package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"

	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) Delete(ctx context.Context, databaseID uuid.UUID) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}
	res, err := m.DatabaseCollection.DeleteOne(ctx, bson.M{"database_id": databaseID, "project_id": projectID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.NotFoundErrorf("database with id %s not found", databaseID)
	}
	return nil
}
