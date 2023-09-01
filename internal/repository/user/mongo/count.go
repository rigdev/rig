package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

// Count returns the amount of user in a project.
func (r *MongoRepository) Count(ctx context.Context) (uint64, error) {
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return 0, err
	}
	mongoFilter := bson.M{"project_id": projectId}
	count, err := r.UsersCol.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return 0, err
	}
	return uint64(count), err
}
