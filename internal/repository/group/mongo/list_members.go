package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/client/mongo"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/internal/repository/group/mongo/schema"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *MongoRepository) ListMembers(ctx context.Context, groupID uuid.UUID, pagination *model.Pagination) (iterator.Iterator[uuid.UUID], uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{}
	filter["project_id"] = projectID
	filter["group_id"] = groupID

	count, err := r.MembersCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := r.MembersCollection.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, 0, err
	}

	it := iterator.NewProducer[uuid.UUID]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var gm schema.GroupMember
			if err := cursor.Decode(&gm); err != nil {
				it.Error(err)
				return
			}
			if err := it.Value(gm.UserID); err != nil {
				it.Error(err)
				return
			}
		}
	}()

	return it, uint64(count), nil
}
