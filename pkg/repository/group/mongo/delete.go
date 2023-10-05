package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"

	"go.mongodb.org/mongo-driver/bson"
)

// Delete removes the group (document) from the database.
func (m *MongoRepository) Delete(ctx context.Context, groupID uuid.UUID) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}
	res, err := m.GroupsCollection.DeleteOne(ctx, bson.M{"group_id": groupID, "project_id": projectID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.NotFoundErrorf("group with id %s not found", groupID)
	}

	filter := bson.M{
		"project_id": projectID,
		"group_id":   groupID,
	}
	if _, err := m.MembersCollection.DeleteMany(ctx, filter); err != nil {
		return err
	}
	return nil
}
