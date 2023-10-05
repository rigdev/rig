package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/pkg/repository/group/mongo/schema"
	"go.mongodb.org/mongo-driver/bson"
)

// Get returns a specific group (document) with "id".
func (m *MongoRepository) Get(ctx context.Context, groupID uuid.UUID) (*group.Group, error) {
	var g schema.Group
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if err := m.GroupsCollection.FindOne(ctx, bson.M{"group_id": groupID, "project_id": projectId}).Decode(&g); err != nil {
		return nil, err
	}
	return g.ToProto()
}
