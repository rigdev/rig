package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/internal/repository/group/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (m *MongoRepository) GetByName(ctx context.Context, name string) (*group.Group, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"project_id": projectID,
		"name":       name,
	}
	g := &schema.Group{}
	if err := m.GroupsCollection.FindOne(ctx, filter).Decode(g); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("group not found")
	} else if err != nil {
		return nil, err
	}

	p, err := g.ToProto()
	return p, err
}
