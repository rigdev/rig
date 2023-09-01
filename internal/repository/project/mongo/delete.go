package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// Delete removes the project (document) from the database.
func (c *MongoRepository) Delete(ctx context.Context, projectID uuid.UUID) error {
	filter := bson.M{"project_id": projectID}
	result, err := c.ProjectCol.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.NotFoundErrorf("project with id %s not found", projectID)
	}
	return nil
}
