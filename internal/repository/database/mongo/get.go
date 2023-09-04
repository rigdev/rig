package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/internal/repository/database/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) Get(ctx context.Context, databaseID uuid.UUID) (*database.Database, uuid.UUID, error) {
	var d schema.Database
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, uuid.Nil, err
	}
	if err := m.DatabaseCollection.FindOne(ctx, bson.M{"database_id": databaseID, "project_id": projectId}).Decode(&d); err != nil {
		return nil, uuid.Nil, err
	}

	db, err := d.ToProto()
	if err != nil {
		return nil, uuid.Nil, err
	}

	return db, d.SecretID, nil
}
