package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/client/mongo"
	"github.com/rigdev/rig/pkg/repository/cluster_config/mongo/schema"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) ListCapsuleConfigs(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*v1alpha1.Capsule], int64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"project_id": projectID,
	}

	count, err := m.CapsuleConfigCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := m.CapsuleConfigCol.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, 0, err
	}

	it := iterator.NewProducer[*v1alpha1.Capsule]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var r schema.CapsuleConfig
			if err := cursor.Decode(&r); err != nil {
				it.Error(err)
				return
			}

			e, err := r.ToAPI()
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

	return it, count, nil
}
