package mongo

import (
	"context"

	"github.com/rigdev/rig/internal/repository/project/mongo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/proto"
)

func (m *MongoRepository) SetSettings(ctx context.Context, projectID string, settingsName string, data []byte) error {
	update := &schema.Settings{
		ProjectID: projectID,
		Name:      settingsName,
		Data:      data,
	}
	_, err := m.SettingsCol.UpdateOne(ctx, bson.M{"project_id": projectID, "name": settingsName}, bson.M{
		"$set": update,
	}, &options.UpdateOptions{Upsert: proto.Bool(true)})
	return err
}
