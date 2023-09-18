package mongo

import (
	"context"

	"github.com/rigdev/rig/internal/repository/cluster_config/mongo/schema"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *MongoRepository) UpdateCapsuleConfig(ctx context.Context, p *v1alpha1.Capsule) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	cp, err := schema.CapsuleConfigFromAPI(projectID, p)
	if err != nil {
		return err
	}

	if err := r.CapsuleConfigCol.FindOneAndUpdate(
		ctx,
		bson.M{
			"project_id": projectID,
			"capsule_id": p.GetName(),
		},
		bson.M{
			"$set": bson.M{
				"data": cp.Data,
			},
		},
	).Decode(&cp); err != nil {
		return err
	}

	return nil
}
