package mongo

import (
	"context"
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func (m *MongoRepository) CreateBuild(ctx context.Context, capsuleID string, b *capsule.Build) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	bp, err := schema.BuildFromProto(projectID, capsuleID, b)
	if err != nil {
		return err
	}

	if _, err := m.BuildCol.InsertOne(ctx, bp); mongo.IsDuplicateKeyError(err) {
		fmt.Println(err)
		return errors.AlreadyExistsErrorf("build already exists")
	} else if err != nil {
		return err
	}

	return nil
}
