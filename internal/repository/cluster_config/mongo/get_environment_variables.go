package mongo

import (
	"context"

	"github.com/rigdev/rig/internal/repository/cluster_config/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (r *MongoRepository) GetEnvironmentVariables(ctx context.Context, capsuleID string) (map[string]string, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	cp := schema.CapsuleConfig{}
	filter := bson.M{"project_id": projectID, "name": capsuleID}
	if err := r.CapsuleConfigCol.FindOne(ctx, filter).Decode(&cp); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("capsule not found")
	} else if err != nil {
		return nil, err
	}

	return cp.EnvironmentVariables, nil
}
