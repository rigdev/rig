package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/internal/repository/service_account/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"go.mongodb.org/mongo-driver/bson"
)

func (t *MongoRepository) List(ctx context.Context) (iterator.Iterator[*service_account.Entry], error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"project_id": projectID,
	}
	cursor, err := t.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	it := iterator.NewProducer[*service_account.Entry]()

	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var sa schema.ServiceAccount
			if err := cursor.Decode(&sa); err != nil {
				it.Error(err)
				return
			}

			p, err := sa.ToProto()
			if err != nil {
				it.Error(err)
				return
			}

			if err := it.Value(&service_account.Entry{
				ServiceAccountId: sa.ServiceAccountID.String(),
				ServiceAccount:   p,
			}); err != nil {
				it.Error(err)
				return
			}
		}
	}()

	return it, nil
}
