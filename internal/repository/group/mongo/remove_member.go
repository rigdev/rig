package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (p *MongoRepository) RemoveMember(ctx context.Context, userID, groupID uuid.UUID) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	filter := bson.M{
		"group_id":   groupID,
		"user_id":    userID,
		"project_id": projectID,
	}

	res, err := p.MembersCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.NotFoundErrorf("group member not found")
	}
	return nil
}
