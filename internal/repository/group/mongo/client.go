// Mongo implements the repository.Group interface.
package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	groupIDIndex = "group_id_idx"
	nameIndex    = "group_name_idx"
	memberIndex  = "group_member_idx"
)

type MongoRepository struct {
	GroupsCollection  *mongo.Collection
	MembersCollection *mongo.Collection
}

func (r *MongoRepository) BuildIndexes(ctx context.Context) error {
	groupIdIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "group_id", Value: 1},
		},
		Options: options.Index().SetName(groupIDIndex).SetUnique(true),
	}
	if _, err := r.GroupsCollection.Indexes().CreateOne(ctx, groupIdIndexModel); err != nil {
		return err
	}
	nameIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "name", Value: 1},
		},
		Options: options.Index().SetName(nameIndex).SetUnique(true),
	}
	if _, err := r.GroupsCollection.Indexes().CreateOne(ctx, nameIndexModel); err != nil {
		return err
	}
	memberIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "group_id", Value: 1},
		},
		Options: options.Index().SetName(memberIndex).SetUnique(true),
	}
	if _, err := r.MembersCollection.Indexes().CreateOne(ctx, memberIndexModel); err != nil {
		return err
	}

	return nil
}

func NewRepository(c *mongo.Client) (*MongoRepository, error) {
	repo := &MongoRepository{
		GroupsCollection:  c.Database("rig").Collection("groups"),
		MembersCollection: c.Database("rig").Collection("groups_to_users"),
	}
	err := repo.BuildIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	return repo, nil
}
