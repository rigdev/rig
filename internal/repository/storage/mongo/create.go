package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/internal/repository/storage/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
)

func (m *MongoRepository) Create(ctx context.Context, secretID uuid.UUID, provider *storage.Provider) (*storage.Provider, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	p, err := schema.ProviderFromProto(projectID, secretID, provider)
	if err != nil {
		return nil, err
	}

	if _, err := m.ProviderCollection.InsertOne(ctx, p); err != nil {
		return nil, err
	}

	return provider, nil
}
