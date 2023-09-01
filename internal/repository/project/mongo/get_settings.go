package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/internal/repository/project/mongo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (m *MongoRepository) GetSettings(ctx context.Context, projectID uuid.UUID, settingsName string) ([]byte, error) {
	resp := schema.Settings{}
	result := m.SettingsCol.FindOne(ctx, bson.M{"project_id": projectID, "name": settingsName})
	if err := result.Err(); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("settings not found")
	} else if err != nil {
		return nil, err
	}
	if err := result.Decode(&resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
