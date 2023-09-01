package mongo

import (
	"context"
	"errors"

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
	err := r.MetricsCol.Database().CreateCollection(
		ctx,
		r.MetricsCol.Name(),
		options.CreateCollection().
			SetTimeSeriesOptions(options.TimeSeries().SetTimeField("timestamp")).
			SetExpireAfterSeconds(metricsExpireAfterSeconds),
	)
	if err != nil {
		var merr mongo.CommandError
		if errors.As(err, &merr) {
			if merr.Code != errorCodeNamespaceExists {
				return err
			}
		} else {
			return err
		}
	}

	capsuleIDIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "capsule_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.CapsuleCol.Indexes().CreateOne(ctx, capsuleIDIndexModel); err != nil {
		return err
	}

	capsuleNameIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "name", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.CapsuleCol.Indexes().CreateOne(ctx, capsuleNameIndexModel); err != nil {
		return err
	}

	rolloutScheduledAtIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "scheduled_at", Value: 1},
		},
		Options: options.Index(),
	}
	if _, err := r.RolloutCol.Indexes().CreateOne(ctx, rolloutScheduledAtIndexModel); err != nil {
		return err
	}

	CapsuleEventsIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "capsule_id", Value: 1},
			{Key: "rollout_id", Value: 1},
		},
		Options: options.Index(),
	}
	if _, err := r.CapsuleEventCol.Indexes().CreateOne(ctx, CapsuleEventsIndexModel); err != nil {
		return err
	}

	metricsIDIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "capsule_id", Value: 1},
			{Key: "instance_id", Value: 1},
		},
		Options: options.Index(),
	}
	if _, err := r.MetricsCol.Indexes().CreateOne(ctx, metricsIDIndexModel); err != nil {
		return err
	}
	return nil
}
