package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) Update(ctx context.Context, providerID uuid.UUID, provider *storage.Provider) (*storage.Provider, error) {
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"provider_id": providerID, "project_id": projectId}
	update := bson.M{"$set": bson.M{"name": provider.Name, "buckets": provider.Buckets}}

	if _, err := m.ProviderCollection.UpdateOne(ctx, filter, update); err != nil {
		return nil, err
	}

	return provider, nil
}
