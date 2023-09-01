package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/internal/repository/storage/mongo/schema"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) LookupByBucket(ctx context.Context, bucket string) (uuid.UUID, *storage.Provider, uuid.UUID, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return uuid.Nil, nil, uuid.Nil, err
	}

	filter := bson.M{"project_id": projectID, "buckets.name": bucket}
	provider := &schema.Provider{}
	if err := m.ProviderCollection.FindOne(ctx, filter).Decode(provider); err != nil {
		return uuid.Nil, nil, uuid.Nil, err
	}

	p, err := provider.ToProto()

	if err != nil {
		return uuid.Nil, nil, uuid.Nil, err
	}

	return provider.ProviderID, p, provider.SecretID, nil
}
