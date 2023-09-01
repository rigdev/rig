package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/internal/repository/storage/mongo/schema"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) List(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*storage.ProviderEntry], uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, 0, err
	}

	count, err := m.ProviderCollection.CountDocuments(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return nil, 0, err
	}

	cursor, err := m.ProviderCollection.Find(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return nil, 0, err
	}

	it := iterator.NewProducer[*storage.ProviderEntry]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var p schema.Provider
			if err := cursor.Decode(&p); err != nil {
				it.Error(err)
				return
			}

			e, err := p.ToProtoEntry()
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
