package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

// Count returns a total count of groups in database.
func (m *MongoRepository) Count(ctx context.Context) (int64, error) {
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return 0, err
	}
	count, err := m.GroupsCollection.CountDocuments(ctx, bson.M{"project_id": projectId})
	if err != nil {
		return 0, err
	}
	return count, nil
}
