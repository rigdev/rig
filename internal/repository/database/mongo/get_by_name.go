package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/internal/repository/database/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) GetByName(ctx context.Context, name string) (*database.Database, uuid.UUID, error) {
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, uuid.Nil, err
	}

	var database *schema.Database
	err = m.DatabaseCollection.FindOne(ctx, bson.M{
		"name":       name,
		"project_id": projectId,
	}).Decode(&database)
	if err != nil {
		return nil, uuid.Nil, err
	}

	db, err := database.ToProto()
	if err != nil {
		return nil, uuid.Nil, err
	}

	return db, database.SecretID, nil
}
