package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/internal/repository/database/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *MongoRepository) GetByName(ctx context.Context, name string) (*database.Database, error) {
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	var database *schema.Database
	err = m.DatabaseCollection.FindOne(ctx, bson.M{
		"name":       name,
		"project_id": projectId,
	}).Decode(&database)
	if err != nil {
		return nil, err
	}

	db, err := database.ToProto()
	if err != nil {
		return nil, err
	}

	return db, nil
}
