package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/internal/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Get returns the requested Project (document) from the database.
func (c *MongoRepository) GetByName(ctx context.Context, name string) (*capsule.Capsule, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	cp := schema.Capsule{}
	if err := c.CapsuleCol.FindOne(ctx, bson.M{
		"project_id": projectID,
		"name":       name,
	}).Decode(&cp); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("capsule not found")
	} else if err != nil {
		return nil, err
	}

	return cp.ToProto()
}
