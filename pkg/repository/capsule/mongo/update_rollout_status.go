package mongo

import (
	"context"

	"github.com/rigdev/rig/gen/go/rollout"
	"github.com/rigdev/rig/pkg/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *MongoRepository) UpdateRolloutStatus(ctx context.Context, capsuleID string, rolloutID uint64, version uint64, rs *rollout.Status) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	rs.Status.UpdatedAt = timestamppb.Now()

	u, err := schema.RolloutStatusFromProto(version, rs)
	if err != nil {
		return err
	}

	r, err := m.RolloutCol.UpdateOne(
		ctx,
		bson.M{
			"project_id": projectID,
			"capsule_id": capsuleID,
			"rollout_id": rolloutID,
			"version":    version,
		},
		u,
	)
	if err != nil {
		return err
	}

	if r.MatchedCount == 0 {
		c, err := m.RolloutCol.CountDocuments(
			ctx,
			bson.M{
				"project_id": projectID,
				"capsule_id": capsuleID,
				"rollout_id": rolloutID,
			},
		)
		if err != nil {
			return err
		}

		if c == 1 {
			return errors.AbortedErrorf("write conflict when updating as version %v", version)
		}
	}

	return nil
}
