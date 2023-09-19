package mongo

import (
	"context"
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/internal/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (m *MongoRepository) GetBuild(ctx context.Context, capsuleID string, buildID string) (*capsule.Build, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	var b schema.Build
	fmt.Println("get build, project", projectID, "capsule", capsuleID, "build", buildID)
	if err := m.BuildCol.FindOne(ctx, bson.M{
		"project_id": projectID,
		"capsule_id": capsuleID,
		"build_id":   buildID,
	}).Decode(&b); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("build not found")
	} else if err != nil {
		return nil, err
	}

	return b.ToProto()
}
