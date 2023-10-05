package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/repository/user/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Update updates the profile info and/or extra data of a specific user.
func (r *MongoRepository) Update(ctx context.Context, p *user.User) (*user.User, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	p.UpdatedAt = timestamppb.Now()
	u, err := schema.UserFromProto(projectID, p)
	if err != nil {
		return nil, err
	}

	if err := r.UsersCol.FindOneAndUpdate(
		ctx,
		schema.GetUserIDFilter(projectID, uuid.UUID(p.GetUserId())),
		bson.M{
			"$set": u,
		},
	).Decode(&u); err != nil {
		return nil, err
	}
	return u.ToProto()
}
