package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/repository/user/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

// Get fetches a user either by id, username, or email.
func (r *MongoRepository) GetByIdentifier(ctx context.Context, id *model.UserIdentifier) (*user.User, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter, err := schema.GetUserIdentifierFilter(projectID, id)
	if err != nil {
		return nil, err
	}
	u := &schema.User{}
	if err := r.UsersCol.FindOne(ctx, filter).Decode(u); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("user not found")
	} else if err != nil {
		return nil, err
	}

	p, err := u.ToProto()
	return p, err
}
