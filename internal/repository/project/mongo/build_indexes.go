package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BuildIndexes TODO: implement.
func (r *MongoRepository) BuildIndexes(ctx context.Context) error {
	projectIdIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.ProjectCol.Indexes().CreateOne(ctx, projectIdIndexModel); err != nil {
		return err
	}
	projectNameIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.ProjectCol.Indexes().CreateOne(ctx, projectNameIndexModel); err != nil {
		return err
	}
	settingsNameIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "name", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.SettingsCol.Indexes().CreateOne(ctx, settingsNameIndexModel); err != nil {
		return err
	}

	return nil
}
