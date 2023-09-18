// Mongo implements the repository.Group interface.
package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository struct {
	CapsuleConfigCol *mongo.Collection
}

func NewRepository(c *mongo.Client) (*MongoRepository, error) {
	db := c.Database("rig")

	repo := &MongoRepository{
		CapsuleConfigCol: db.Collection("capsule_config"),
	}

	err := repo.BuildIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	return repo, nil
}
