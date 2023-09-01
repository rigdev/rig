package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/internal/repository/user/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

// Create inserts a user in the database.
func (r *MongoRepository) Create(ctx context.Context, p *user.User) (*user.User, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	u, err := schema.UserFromProto(projectID, p)
	if err != nil {
		return nil, err
	}

	if _, err := r.UsersCol.InsertOne(ctx, u); mongo.IsDuplicateKeyError(err) {
		return nil, errors.AlreadyExistsErrorf("user-identifier already exists")
	} else if err != nil {
		return nil, err
	}
	// set new data for user created
	return p, nil
}
