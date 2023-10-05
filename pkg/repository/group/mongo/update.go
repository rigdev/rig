package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/repository/group/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Update updates the name of a specific group.
func (m *MongoRepository) Update(ctx context.Context, group *group.Group) (*group.Group, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}
	group.UpdatedAt = timestamppb.Now()
	g, err := schema.GroupFromProto(projectID, group)
	if err != nil {
		return nil, err
	}
	if err := m.GroupsCollection.FindOneAndUpdate(
		ctx,
		schema.GetGroupIDFilter(projectID, group.GetGroupId()),
		bson.M{
			"$set": g,
		},
	).Decode(&g); err != nil {
		return nil, err
	}

	return g.ToProto()
}
