package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

// DeleteAll deletes all users in the collection.
func (r *MongoRepository) DeleteAll(ctx context.Context) error {
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	filter := bson.M{"project_id": projectId}

	if _, err := r.UsersCol.DeleteMany(ctx, filter); err != nil {
		return err
	}
	return nil
}
