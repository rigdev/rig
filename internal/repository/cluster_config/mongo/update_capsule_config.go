package mongo

import (
	"context"

	"github.com/rigdev/rig/gen/go/capsule"
	"github.com/rigdev/rig/internal/repository/cluster_config/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *MongoRepository) UpdateCapsuleConfig(ctx context.Context, p *capsule.Config) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	cp, err := schema.CapsuleConfigFromProto(projectID, p)
	if err != nil {
		return err
	}

	if err := r.CapsuleConfigCol.FindOneAndUpdate(
		ctx,
		bson.M{
			"project_id": projectID,
			"name":       p.GetName(),
		},
		bson.M{
			"$set": cp,
		},
	).Decode(&cp); err != nil {
		return err
	}

	return nil
}
