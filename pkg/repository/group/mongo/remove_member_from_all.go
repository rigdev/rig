package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (p *MongoRepository) RemoveMemberFromAll(ctx context.Context, userID uuid.UUID) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}
	filter := bson.M{
		"project_id": projectID,
		"user_id":    userID,
	}
	if _, err := p.MembersCollection.DeleteMany(ctx, filter); err != nil {
		return err
	}
	return nil
}
