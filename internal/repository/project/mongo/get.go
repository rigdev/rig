package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/internal/repository/project/mongo/schema"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Get returns the requested Project (document) from the database.
func (c *MongoRepository) Get(ctx context.Context, projectID string) (*project.Project, error) {
	resp := schema.Project{}
	result := c.ProjectCol.FindOne(ctx, bson.M{"project_id": projectID})
	if err := result.Err(); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("project not found")
	} else if err != nil {
		return nil, err
	}
	if err := result.Decode(&resp); err != nil {
		return nil, err
	}
	return resp.ToProto()
}
