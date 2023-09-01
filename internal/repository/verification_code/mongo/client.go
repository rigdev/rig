package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	expiresAfterIndex = "expires_after_index"
)

type MongoRepository struct {
	Collection *mongo.Collection
}

func NewRepository(c *mongo.Client) (*MongoRepository, error) {
	repo := &MongoRepository{
		Collection: c.Database("rig").Collection("verification_codes"),
	}
	err := repo.BuildIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// BuildIndexes builds the indexes for the verification repo.
func (r *MongoRepository) BuildIndexes(ctx context.Context) error {
	expiresAtIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "expires_at", Value: 1},
		},
		Options: options.Index().SetExpireAfterSeconds(0).SetName(expiresAfterIndex),
	}
	if _, err := r.Collection.Indexes().CreateOne(ctx, expiresAtIndexModel); err != nil {
		return err
	}
	typeUserIdCombinedIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "type", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.Collection.Indexes().CreateOne(ctx, typeUserIdCombinedIndexModel); err != nil {
		return err
	}

	projectIdIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
		},
		Options: options.Index().SetName("project_id_index"),
	}
	if _, err := r.Collection.Indexes().CreateOne(ctx, projectIdIndexModel); err != nil {
		return err
	}
	return nil
}
