package mongo

import (
	"context"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// IncreaseAttempsts increases the verifcation attempt by 1.
func (r *MongoRepository) IncreaseAttempts(ctx context.Context, userID uuid.UUID, t user.VerificationType) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	mongoUpdate := bson.M{
		"$set": bson.M{
			"last_attempt": time.Now(),
		},
		"$inc": bson.D{{Key: "attempts", Value: 1}},
	}
	_, err = r.Collection.UpdateOne(ctx, bson.M{"project_id": projectID, "user_id": userID, "verification_type": t}, mongoUpdate)
	if err != nil {
		return err
	}
	return nil
}
