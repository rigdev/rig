package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	maximumGetLimit = 75
	userIDIndex     = "user_id_idx"
	emailIndex      = "user_email_idx"
	usernameIndex   = "user_username_idx"
	phoneIndex      = "user_phone_idx"
	oauth2Index     = "oauth2_idx"
)

type MongoRepository struct {
	UsersCol         *mongo.Collection
	Oauth2Col        *mongo.Collection
	ValidatePassword bool
}

func NewRepository(c *mongo.Client) (*MongoRepository, error) {
	repo := &MongoRepository{
		UsersCol:  c.Database("rig").Collection("users"),
		Oauth2Col: c.Database("rig").Collection("oauth2"),
	}
	err := repo.BuildIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// BuildIndexes builds the MongoDB indexes for the user object.
func (r *MongoRepository) BuildIndexes(ctx context.Context) error {
	// Users collection indices
	userIDIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "user_id", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName(userIDIndex),
	}
	if _, err := r.UsersCol.Indexes().CreateOne(ctx, userIDIndexModel); err != nil {
		return err
	}
	emailProjectIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "email", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(
			bson.D{
				{
					Key: "email", Value: bson.D{
						{
							Key: "$gt", Value: "",
						},
					},
				},
			},
		).SetName(emailIndex),
	}
	if _, err := r.UsersCol.Indexes().CreateOne(ctx, emailProjectIndexModel); err != nil {
		return err
	}
	usernameIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "username", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(
			bson.D{
				{
					Key: "username", Value: bson.D{
						{
							Key: "$gt", Value: "",
						},
					},
				},
			},
		).SetName(usernameIndex),
	}
	if _, err := r.UsersCol.Indexes().CreateOne(ctx, usernameIndexModel); err != nil {
		return err
	}
	phoneIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "phone_number", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetPartialFilterExpression(
			bson.D{
				{
					Key: "phone_number", Value: bson.D{
						{
							Key: "$gt", Value: "",
						},
					},
				},
			},
		).SetName(phoneIndex),
	}
	if _, err := r.UsersCol.Indexes().CreateOne(ctx, phoneIndexModel); err != nil {
		return err
	}

	// Oauth2 collection indices
	oauth2IndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "sub", Value: 1},
			{Key: "iss", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName(oauth2Index),
	}
	if _, err := r.Oauth2Col.Indexes().CreateOne(ctx, oauth2IndexModel); err != nil {
		return err
	}

	return nil
}
