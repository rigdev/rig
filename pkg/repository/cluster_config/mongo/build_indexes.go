package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	metricsExpireAfterSeconds = 60 * 15 // TODO: move to config
	errorCodeNamespaceExists  = 48
)

// BuildIndexes TODO: implement.
func (r *MongoRepository) BuildIndexes(ctx context.Context) error {
	capsuleConfigIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "capsule_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.CapsuleConfigCol.Indexes().CreateOne(ctx, capsuleConfigIndexModel); err != nil {
		return err
	}

	return nil
}
