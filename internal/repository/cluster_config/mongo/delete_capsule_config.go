package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func (c *MongoRepository) DeleteCapsuleConfig(ctx context.Context, capsuleName string) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	filter := bson.M{"project_id": projectID, "name": capsuleName}
	result, err := c.CapsuleConfigCol.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.NotFoundErrorf("capsule not found")
	}
	return nil
}
