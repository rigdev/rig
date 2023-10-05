package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/gen/go/rollout"
	"github.com/rigdev/rig/pkg/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoRepository) CreateRollout(ctx context.Context, capsuleID string, rc *capsule.RolloutConfig, rs *rollout.Status) (uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return 0, err
	}

	filter := bson.M{"project_id": projectID, "capsule_id": capsuleID}
	options := options.FindOne()
	options.SetSort(bson.D{{Key: "rollout_id", Value: -1}})

	var rolloutID uint64
	var r schema.Rollout
	if err := m.RolloutCol.FindOne(ctx, filter, options).Decode(&r); err == mongo.ErrNoDocuments {
		// None exists.
	} else if err != nil {
		return 0, err
	} else {
		rolloutID = r.RolloutID
	}

	rolloutID++

	rp, err := schema.RolloutFromProto(projectID, capsuleID, rolloutID, 1, rc, rs)
	if err != nil {
		return 0, err
	}

	if _, err := m.RolloutCol.InsertOne(ctx, rp); mongo.IsDuplicateKeyError(err) {
		return 0, errors.AlreadyExistsErrorf("rollout already exists")
	} else if err != nil {
		return 0, err
	}

	return rolloutID, nil
}
