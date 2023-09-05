package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

// Update updates the name of a specific group.
func (m *MongoRepository) Update(ctx context.Context, database *database.Database) (*database.Database, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"database_id": database.GetDatabaseId(), "project_id": projectID}
	update := bson.M{"$set": bson.M{"name": database.GetName(), "tables": database.GetTables(), "client_ids": database.GetClientIds()}}

	if _, err := m.DatabaseCollection.UpdateOne(ctx, filter, update); err != nil {
		return nil, err
	}

	return database, nil
}
