package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/client/mongo"
	"github.com/rigdev/rig/internal/repository/group/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
)

// List implements group_repository.GroupRepository
func (m *MongoRepository) List(ctx context.Context, pagination *model.Pagination, search string) (iterator.Iterator[*group.Group], uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := mongo.StringSearch(search)
	filter["project_id"] = projectID

	count, err := m.GroupsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := m.GroupsCollection.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, 0, err
	}

	it := iterator.NewProducer[*group.Group]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var g schema.Group
			if err := cursor.Decode(&g); err != nil {
				it.Error(err)
				return
			}

			e, err := g.ToProto()
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
