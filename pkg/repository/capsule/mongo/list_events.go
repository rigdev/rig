package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/client/mongo"
	"github.com/rigdev/rig/pkg/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) ListEvents(ctx context.Context, pagination *model.Pagination, capsuleID string, rolloutID uint64) (iterator.Iterator[*capsule.Event], uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"project_id": projectID,
		"capsule_id": capsuleID,
		"rollout_id": rolloutID,
	}

	count, err := m.CapsuleEventCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := m.CapsuleEventCol.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, 0, err
	}

	it := iterator.NewProducer[*capsule.Event]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var e schema.Event
			if err := cursor.Decode(&e); err != nil {
				it.Error(err)
				return
			}

			p, err := e.ToProto()
			if err != nil {
				it.Error(err)
				return
			}

			if err := it.Value(p); err != nil {
				it.Error(err)
				return
			}
		}
	}()

	return it, uint64(count), nil
}
