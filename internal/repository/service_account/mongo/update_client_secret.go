package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
)

// UpdatePassword hashes and updates the password of a specific user.
func (r *MongoRepository) UpdateClientSecret(ctx context.Context, credentialID uuid.UUID, pw *model.HashingInstance) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	bs, err := proto.Marshal(pw)
	if err != nil {
		return err
	}

	return r.Collection.FindOneAndUpdate(
		ctx,
		bson.M{
			"project_id":         projectID,
			"service_account_id": credentialID,
		},
		bson.M{
			"$set": bson.M{
				"password": bs,
			},
		},
	).Err()
}
