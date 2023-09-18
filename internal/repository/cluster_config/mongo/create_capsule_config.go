package mongo

import (
	"context"

	"github.com/rigdev/rig/internal/repository/cluster_config/mongo/schema"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *MongoRepository) CreateCapsuleConfig(ctx context.Context, p *v1alpha1.Capsule) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	cp, err := schema.CapsuleConfigFromAPI(projectID, p)
	if err != nil {
		return err
	}

	// insert in database
	if _, err := c.CapsuleConfigCol.InsertOne(ctx, cp); mongo.IsDuplicateKeyError(err) {
		return errors.AlreadyExistsErrorf("capsule config already exists")
	} else if err != nil {
		return err
	}

	return nil
}
