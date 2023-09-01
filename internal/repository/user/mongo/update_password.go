package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/internal/repository/user/mongo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
)

// UpdatePassword hashes and updates the password of a specific user.
func (r *MongoRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, password *model.HashingInstance) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	bs, err := proto.Marshal(password)
	if err != nil {
		return err
	}

	return r.UsersCol.FindOneAndUpdate(
		ctx,
		schema.GetUserIDFilter(projectID, userID),
		bson.M{
			"$set": bson.M{
				"password": bs,
			},
		},
	).Err()
}
