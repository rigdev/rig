package mongo

import (
	"context"

	"github.com/rigdev/rig/internal/repository/cluster_config/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	v1 "k8s.io/api/core/v1"
)

func (r *MongoRepository) GetFiles(ctx context.Context, capsuleID string) ([]*v1.ConfigMap, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	cp := schema.CapsuleConfig{}
	filter := bson.M{"project_id": projectID, "capsule_id": capsuleID}
	if err := r.CapsuleConfigCol.FindOne(ctx, filter).Decode(&cp); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("capsule not found")
	} else if err != nil {
		return nil, err
	}

	return cp.Files, nil
}
