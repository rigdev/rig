package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/pkg/repository/project/mongo/schema"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

// Create inserts a model.Project in the database with default fields and returns the created object.
func (c *MongoRepository) Create(ctx context.Context, project *project.Project) (*project.Project, error) {
	// convert to standard model
	p, err := schema.ProjectFromProto(project)
	if err != nil {
		return nil, err
	}
	// insert in database
	if _, err := c.ProjectCol.InsertOne(ctx, p); mongo.IsDuplicateKeyError(err) {
		return nil, errors.AlreadyExistsErrorf("project already exists")
	} else if err != nil {
		return nil, err
	}
	return project, nil
}
