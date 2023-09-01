package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/internal/repository/user/mongo/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

// Get fetches a user either by id, username, or email.
func (r *MongoRepository) Get(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := schema.GetUserIDFilter(projectID, userID)
	u := schema.User{}
	if err := r.UsersCol.FindOne(ctx, filter).Decode(&u); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("user not found")
	} else if err != nil {
		return nil, err
	}

	return u.ToProto()
}
