package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/proto"
)

// Create inserts a new verification attempt in the database.
func (c *MongoRepository) Create(ctx context.Context, vc *user.VerificationCode) (*user.VerificationCode, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	bs, err := proto.Marshal(vc)
	if err != nil {
		return nil, err
	}

	_, err = c.Collection.UpdateOne(ctx,
		bson.M{"project_id": projectID, "user_id": vc.GetUserId(), "verification_type": vc.GetType()},
		bson.M{"$set": bson.M{
			"data":     bs,
			"attempts": vc.GetAttempts(),
		}}, options.Update().SetUpsert(true))
	if err != nil {
		return nil, err
	}

	return vc, nil
}
