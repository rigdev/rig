package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/pkg/repository/user/mongo/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

// Delete deletes a specific user with id, email or username.
func (r *MongoRepository) Delete(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := schema.GetUserIDFilter(projectID, userID)
	var resp schema.User
	if err := r.UsersCol.FindOneAndDelete(ctx, filter).Decode(&resp); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("user not found")
	} else if err != nil {
		return nil, err
	}
	return resp.ToProto()
}
