package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/internal/repository/storage/mongo/schema"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) Get(ctx context.Context, providerID uuid.UUID) (*storage.Provider, uuid.UUID, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, uuid.Nil, err
	}

	filter := bson.M{"provider_id": providerID, "project_id": projectID}
	var p schema.Provider
	if err := m.ProviderCollection.FindOne(ctx, filter).Decode(&p); err != nil {
		return nil, uuid.Nil, err
	}

	provider, err := p.ToProto()
	if err != nil {
		return nil, uuid.Nil, err
	}

	return provider, p.SecretID, nil
}
