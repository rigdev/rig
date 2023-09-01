package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// Delete removes a verfication attempt from the database.
func (c *MongoRepository) Delete(ctx context.Context, userID uuid.UUID, t user.VerificationType) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	result, err := c.Collection.DeleteOne(ctx, bson.M{"project_id": projectID, "user_id": userID, "verification_type": t})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.NotFoundErrorf("verification code not found")
	}
	return nil
}
