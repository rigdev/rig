package mongo

import (
	"context"

	"github.com/rigdev/rig/gen/go/capsule"
	"github.com/rigdev/rig/internal/repository/cluster_config/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Get returns the requested Project (document) from the database.
func (c *MongoRepository) GetCapsuleConfig(ctx context.Context, capsuleName string) (*capsule.Config, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	cp := schema.CapsuleConfig{}
	filter := bson.M{"project_id": projectID, "name": capsuleName}
	if err := c.CapsuleConfigCol.FindOne(ctx, filter).Decode(&cp); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("capsule not found")
	} else if err != nil {
		return nil, err
	}

	return cp.ToProto()
}
