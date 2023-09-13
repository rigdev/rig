package mongo

import (
	"context"

	"github.com/rigdev/rig/gen/go/capsule"
	"github.com/rigdev/rig/internal/repository/cluster_config/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *MongoRepository) CreateCapsuleConfig(ctx context.Context, p *capsule.Config) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	cp, err := schema.CapsuleConfigFromProto(projectID, p)
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
