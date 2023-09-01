package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/internal/repository/session/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// GetTokens fetches all tokens for a specific user.
func (r *MongoRepository) List(ctx context.Context, userID uuid.UUID) (iterator.Iterator[*user.SessionEntry], error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"project_id": projectID, "user_id": userID}
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	it := iterator.NewProducer[*user.SessionEntry]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var s schema.Session
			if err := cursor.Decode(&s); err != nil {
				it.Error(err)
				return
			}

			p, err := s.ToProto()
			if err != nil {
				it.Error(err)
				return
			}

			if err := it.Value(&user.SessionEntry{
				SessionId: s.SessionID.String(),
				Session:   p,
			}); err != nil {
				it.Error(err)
				return
			}
		}
	}()

	return it, nil
}
