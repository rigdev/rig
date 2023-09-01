package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/internal/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *MongoRepository) Create(ctx context.Context, p *capsule.Capsule) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	cp, err := schema.CapsuleFromProto(projectID, p)
	if err != nil {
		return err
	}
	// insert in database
	if _, err := c.CapsuleCol.InsertOne(ctx, cp); mongo.IsDuplicateKeyError(err) {
		return errors.AlreadyExistsErrorf("capsule already exists")
	} else if err != nil {
		return err
	}

	return nil
}
