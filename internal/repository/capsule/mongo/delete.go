package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

// Delete removes the project (document) from the database.
func (c *MongoRepository) Delete(ctx context.Context, capsuleID string) error {
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

	if _, err := c.CapsuleEventCol.DeleteMany(ctx, filter); err != nil {
		return err
	}

	if _, err := c.MetricsCol.DeleteMany(ctx, filter); err != nil {
		return err
	}

	return nil
}
