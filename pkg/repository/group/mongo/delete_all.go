package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) DeleteAll(ctx context.Context) error {
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	filter := bson.M{"project_id": projectId}

	if _, err := m.GroupsCollection.DeleteMany(ctx, filter); err != nil {
		return err
	}

	return nil
}
