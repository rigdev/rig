package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *MongoRepository) CountMembers(ctx context.Context, groupID uuid.UUID) (uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return 0, err
	}
	filter := bson.M{
		"project_id": projectID,
		"group_id":   groupID,
	}

	count, err := r.MembersCollection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return uint64(count), nil
}
