package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
	v1 "k8s.io/api/core/v1"
)

func (r *MongoRepository) SetFiles(ctx context.Context, capsuleID string, files []*v1.ConfigMap) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	if err := r.CapsuleConfigCol.FindOneAndUpdate(
		ctx,
		bson.M{
			"project_id": projectID,
			"capsule_id": capsuleID,
		},
		bson.M{
			"$set": bson.M{
				"files": files,
			},
		},
	).Err(); err != nil {
		return err
	}

	return nil
}
