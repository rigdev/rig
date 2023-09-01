package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	providerIDIndex   = "provider_id_idx"
	providerNameIndex = "provider_name_idx"
	bucketNameIndex   = "bucket_name_idx"
)

type MongoRepository struct {
	ProviderCollection *mongo.Collection
}

func NewRepository(c *mongo.Client) (*MongoRepository, error) {
	repo := &MongoRepository{
		ProviderCollection: c.Database("rig").Collection("providers"),
	}
	err := repo.BuildIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *MongoRepository) BuildIndexes(ctx context.Context) error {
	providerIDIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "provider_id", Value: 1},
		},
		Options: options.Index().SetName(providerIDIndex).SetUnique(true),
	}
	if _, err := r.ProviderCollection.Indexes().CreateOne(ctx, providerIDIndexModel); err != nil {
		return err
	}

	providerNameIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "name", Value: 1},
		},
		Options: options.Index().SetName(providerNameIndex).SetUnique(true),
	}
	if _, err := r.ProviderCollection.Indexes().CreateOne(ctx, providerNameIndexModel); err != nil {
		return err
	}

	bucketNameIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "buckets.name", Value: 1},
		},
		Options: options.Index().SetName(bucketNameIndex).SetUnique(true).SetPartialFilterExpression(bson.M{"buckets.name": bson.M{"$exists": true}}),
	}
	if _, err := r.ProviderCollection.Indexes().CreateOne(ctx, bucketNameIndexModel); err != nil {
		return err
	}
	return nil
}
