package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/internal/repository/database/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
)

func (m *MongoRepository) Create(ctx context.Context, secretID uuid.UUID, database *database.Database) (*database.Database, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}
	d, err := schema.DatabaseFromProto(projectID, secretID, database)
	if err != nil {
		return nil, err
	}
	if _, err := m.DatabaseCollection.InsertOne(ctx, d); err != nil {
		return nil, err
	}
	return database, nil
}
