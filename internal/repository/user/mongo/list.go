package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/client/mongo"
	"github.com/rigdev/rig/internal/repository/user/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
)

// List fetches all users matching the filter.
func (r *MongoRepository) List(ctx context.Context, set *settings.Settings, pagination *model.Pagination, search string) (iterator.Iterator[*model.UserEntry], uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := mongo.StringSearch(search)
	filter["project_id"] = projectID

	count, err := r.UsersCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := r.UsersCol.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, 0, err
	}

	it := iterator.NewProducer[*model.UserEntry]()

	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var u schema.User
			if err := cursor.Decode(&u); err != nil {
				it.Error(err)
				return
			}

			e, err := u.ToProtoEntry(set)
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
