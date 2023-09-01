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

func (r *MongoRepository) List(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.Capsule], int64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"project_id": projectID,
	}

	count, err := r.CapsuleCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := r.CapsuleCol.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, 0, err
	}

	it := iterator.NewProducer[*capsule.Capsule]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var cp schema.Capsule
			if err := cursor.Decode(&cp); err != nil {
				it.Error(err)
				return
			}

			e, err := cp.ToProto()
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
