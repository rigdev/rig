package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	serviceAccountIDIndex = "service_account_id_index"
	nameIndex             = "name_index"
)

type MongoRepository struct {
	Collection *mongo.Collection
}

func NewRepository(c *mongo.Client) (*MongoRepository, error) {
	repo := &MongoRepository{
		Collection: c.Database("rig").Collection("service_accounts"),
	}
	err := repo.BuildIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// BuildIndexes builds the indexes for the Mongo Object.
func (r *MongoRepository) BuildIndexes(ctx context.Context) error {
	credentialIDIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "service_account_id", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName(serviceAccountIDIndex),
	}
	if _, err := r.Collection.Indexes().CreateOne(ctx, credentialIDIndexModel); err != nil {
		return err
	}
	userIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "name", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName(nameIndex),
	}
	if _, err := r.Collection.Indexes().CreateOne(ctx, userIndexModel); err != nil {
		return err
	}
	return nil
}
