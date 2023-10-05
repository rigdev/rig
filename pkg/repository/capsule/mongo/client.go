// Mongo implements the repository.Group interface.
package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository struct {
	BuildCol        *mongo.Collection
	RolloutCol      *mongo.Collection
	CapsuleEventCol *mongo.Collection
	MetricsCol      *mongo.Collection
}

func NewRepository(c *mongo.Client) (*MongoRepository, error) {
	db := c.Database("rig")

	repo := &MongoRepository{
		BuildCol:        c.Database("rig").Collection("builds"),
		RolloutCol:      c.Database("rig").Collection("rollouts"),
		CapsuleEventCol: c.Database("rig").Collection("capsule_events"),
		MetricsCol:      db.Collection("capsule_metrics"),
	}

	err := repo.BuildIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	return repo, nil
}
