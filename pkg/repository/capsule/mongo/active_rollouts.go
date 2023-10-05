package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/client/mongo"
	repo_capsule "github.com/rigdev/rig/pkg/repository/capsule"
	"github.com/rigdev/rig/pkg/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/iterator"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) ActiveRollouts(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[repo_capsule.ActiveRollout], error) {
	filter := bson.M{
		"scheduled_at": bson.M{
			"$exists": true,
		},
	}

	cursor, err := m.RolloutCol.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, err
	}

	it := iterator.NewProducer[repo_capsule.ActiveRollout]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var r schema.Rollout
			if err := cursor.Decode(&r); err != nil {
				it.Error(err)
				return
			}

			ar := repo_capsule.ActiveRollout{
				ProjectID: r.ProjectID,
				CapsuleID: r.CapsuleID,
				RolloutID: r.RolloutID,
			}
			if r.ScheduledAt != nil {
				ar.ScheduledAt = *r.ScheduledAt
			}

			if err := it.Value(ar); err != nil {
				it.Error(err)
				return
			}
		}
	}()

	return it, nil
}
