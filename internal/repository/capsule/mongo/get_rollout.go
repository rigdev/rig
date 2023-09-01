package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/gen/go/rollout"
	"github.com/rigdev/rig/internal/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (m *MongoRepository) GetRollout(ctx context.Context, capsuleID uuid.UUID, rolloutID uint64) (*capsule.RolloutConfig, *rollout.Status, uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	filter := bson.M{"project_id": projectID, "capsule_id": capsuleID, "rollout_id": rolloutID}
	var r schema.Rollout
	if err := m.RolloutCol.FindOne(ctx, filter).Decode(&r); err == mongo.ErrNoDocuments {
		return nil, nil, 0, errors.NotFoundErrorf("rollout not found")
	} else if err != nil {
		return nil, nil, 0, err
	}

	rc, err := r.ConfigToProto()
	if err != nil {
		return nil, nil, 0, err
	}

	rs, err := r.StatusToProto()
	if err != nil {
		return nil, nil, 0, err
	}

	return rc, rs, r.Version, err
}
