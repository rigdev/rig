package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/internal/repository/service_account/mongo/schema"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// Get fetches a specific token from the database.
func (t *MongoRepository) Get(ctx context.Context, serviceAccountID uuid.UUID) (string, *service_account.ServiceAccount, error) {
	filter := &bson.M{
		"service_account_id": serviceAccountID,
	}
	c := schema.ServiceAccount{}
	if err := t.Collection.FindOne(ctx, filter).Decode(&c); err != nil {
		return "", nil, err
	}

	p, err := c.ToProto()
	return c.ProjectID, p, err
}
