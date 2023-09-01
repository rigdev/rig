package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// Delete removes the project (document) from the database.
func (c *MongoRepository) Delete(ctx context.Context, capsuleID uuid.UUID) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	filter := bson.M{"project_id": projectID, "capsule_id": capsuleID}
	if _, err := c.BuildCol.DeleteMany(ctx, filter); err != nil {
		return err
	}

	if _, err := c.RolloutCol.DeleteMany(ctx, filter); err != nil {
		return err
	}

	result, err := c.CapsuleCol.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.NotFoundErrorf("capsule not found")
	}
	return nil
}
