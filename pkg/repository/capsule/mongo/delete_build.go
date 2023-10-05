package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) DeleteBuild(ctx context.Context, capsuleID string, buildID string) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	filter := bson.M{"project_id": projectID, "capsule_id": capsuleID, "build_id": buildID}
	result, err := m.BuildCol.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.NotFoundErrorf("build not found")
	}
	return nil
}
