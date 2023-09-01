package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/internal/repository/database/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

// Update updates the name of a specific group.
func (m *MongoRepository) Update(ctx context.Context, database *database.Database) (*database.Database, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}
	d, err := schema.DatabaseFromProto(projectID, database)
	if err != nil {
		return nil, err
	}
	if err := m.DatabaseCollection.FindOneAndUpdate(
		ctx,
		bson.M{"project_id": projectID, "database_id": database.GetDatabaseId()},
		bson.M{
			"$set": d,
		},
	).Decode(&d); err != nil {
		return nil, err
	}

	return d.ToProto()
}
