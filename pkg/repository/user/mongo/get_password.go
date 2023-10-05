package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/pkg/repository/user/mongo/schema"
	"google.golang.org/protobuf/proto"
)

// Get fetches a user either by id, username, or email.
func (r *MongoRepository) GetPassword(ctx context.Context, userID uuid.UUID) (*model.HashingInstance, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := schema.GetUserIDFilter(projectID, userID)
	var d struct {
		Password []byte
	}
	if err := r.UsersCol.FindOne(ctx, filter).Decode(&d); err != nil {
		return nil, err
	}

	pw := &model.HashingInstance{}
	return pw, proto.Unmarshal(d.Password, pw)
}
