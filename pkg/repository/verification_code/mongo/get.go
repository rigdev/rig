package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/proto"
)

// Get fetches a verfication session from the database.
func (r *MongoRepository) Get(ctx context.Context, userID uuid.UUID, verificationType user.VerificationType) (*user.VerificationCode, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data     []byte
		Attempts int32
	}
	result := r.Collection.FindOne(ctx, bson.M{"project_id": projectID, "user_id": userID, "verification_type": verificationType})
	if err := result.Err(); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("verification code not found")
	} else if err != nil {
		return nil, err
	}
	if err := result.Decode(&resp); err != nil {
		return nil, err
	}

	vc := &user.VerificationCode{}
	if err := proto.Unmarshal(resp.Data, vc); err != nil {
		return nil, err
	}

	vc.Attempts = resp.Attempts

	return vc, nil
}
