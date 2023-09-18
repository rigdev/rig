package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
	v1 "k8s.io/api/core/v1"
)

func (r *MongoRepository) SetSecrets(ctx context.Context, capsuleID string, secrets []*v1.Secret) error {
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
				"secrets": secrets,
			},
		},
	).Err(); err != nil {
		return err
	}

	return nil
}
