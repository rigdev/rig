package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/client/mongo"
	"github.com/rigdev/rig/internal/repository/database/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"go.mongodb.org/mongo-driver/bson"
)

// List implements group_repository.GroupRepository
func (m *MongoRepository) List(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*database.Database], uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"project_id": projectID}

	count, err := m.DatabaseCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := m.DatabaseCollection.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, 0, err
	}

	it := iterator.NewProducer[*database.Database]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var d schema.Database
			if err := cursor.Decode(&d); err != nil {
				it.Error(err)
				return
			}

			e, err := d.ToProto()
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
