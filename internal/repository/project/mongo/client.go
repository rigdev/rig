// Mongo implements the repository.Group interface.
package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository struct {
	ProjectCol  *mongo.Collection
	SettingsCol *mongo.Collection
}

func NewRepository(c *mongo.Client) (*MongoRepository, error) {
	repo := &MongoRepository{
		ProjectCol:  c.Database("rig").Collection("projects"),
		SettingsCol: c.Database("rig").Collection("settings"),
	}
	err := repo.BuildIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	return repo, nil
}
