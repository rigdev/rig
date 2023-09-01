package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/internal/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *MongoRepository) Update(ctx context.Context, p *capsule.Capsule) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	cp, err := schema.CapsuleFromProto(projectID, p)
	if err != nil {
		return err
	}

	if err := r.CapsuleCol.FindOneAndUpdate(
		ctx,
		bson.M{
			"project_id": projectID,
			"capsule_id": p.GetCapsuleId(),
		},
		bson.M{
			"$set": cp,
		},
	).Decode(&cp); err != nil {
		return err
	}

	return nil
}
