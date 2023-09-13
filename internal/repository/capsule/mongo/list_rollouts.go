package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/client/mongo"
	"github.com/rigdev/rig/internal/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) ListRollouts(ctx context.Context, pagination *model.Pagination, capsuleID string) (iterator.Iterator[*capsule.Rollout], uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"project_id": projectID,
		"capsule_id": capsuleID,
	}

	count, err := m.RolloutCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := m.RolloutCol.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, 0, err
	}

	it := iterator.NewProducer[*capsule.Rollout]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var r schema.Rollout
			if err := cursor.Decode(&r); err != nil {
				it.Error(err)
				return
			}

			e, err := r.ToProto()
			if err != nil {
				it.Error(err)
				return
			}

			if err := it.Value(e); err != nil {
				it.Error(err)
				return
			}
		}
	}()

	return it, uint64(count), nil
}
