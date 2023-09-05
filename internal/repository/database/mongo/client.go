package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	databaseIDIndex = "database_id_idx"
	nameIndex       = "database_name_idx"
	tableNameIndex  = "database_table_name_idx"
	clientIdIndex   = "database_client_id_idx"
)

type MongoRepository struct {
	DatabaseCollection *mongo.Collection
}

func (r *MongoRepository) BuildIndexes(ctx context.Context) error {
	databaseIdIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "database_id", Value: 1},
		},
		Options: options.Index().SetName(databaseIDIndex).SetUnique(true),
	}
	if _, err := r.DatabaseCollection.Indexes().CreateOne(ctx, databaseIdIndexModel); err != nil {
		return err
	}
	nameIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "name", Value: 1},
		},
		Options: options.Index().SetName(nameIndex).SetUnique(true),
	}
	if _, err := r.DatabaseCollection.Indexes().CreateOne(ctx, nameIndexModel); err != nil {
		return err
	}
	tableNameIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "tables.name", Value: 1},
		},
		Options: options.Index().SetName(tableNameIndex).SetUnique(true).SetPartialFilterExpression(bson.M{"tables.name": bson.M{"$exists": true}}),
	}
	if _, err := r.DatabaseCollection.Indexes().CreateOne(ctx, tableNameIndexModel); err != nil {
		return err
	}
	clientIdIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "client_ids", Value: 1},
		},
		Options: options.Index().SetName(clientIdIndex).SetUnique(true).SetPartialFilterExpression(bson.M{"client_ids": bson.M{"$exists": true}}),
	}
	if _, err := r.DatabaseCollection.Indexes().CreateOne(ctx, clientIdIndexModel); err != nil {
		return err
	}
	return nil
}

func NewRepository(c *mongo.Client) (*MongoRepository, error) {
	repo := &MongoRepository{
		DatabaseCollection: c.Database("rig").Collection("databases"),
	}
	err := repo.BuildIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	return repo, nil
}
