package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/internal/repository/group/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
)

// Create inserts a new group (document) in the database.
func (m *MongoRepository) Create(ctx context.Context, group *group.Group) (*group.Group, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}
	g, err := schema.GroupFromProto(projectID, group)
	if err != nil {
		return nil, err
	}
	if _, err := m.GroupsCollection.InsertOne(ctx, g); err != nil {
		return nil, err
	}
	return group, nil
}
