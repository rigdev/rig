package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/repository/service_account/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

func (r *MongoRepository) GetClientSecret(ctx context.Context, serviceAccountID uuid.UUID) (*model.HashingInstance, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := schema.GetServiceAccountIDFilter(projectID, serviceAccountID)
	var d struct {
		Password []byte
	}
	if err := r.Collection.FindOne(ctx, filter).Decode(&d); err != nil {
		return nil, err
	}

	pw := &model.HashingInstance{}
	return pw, proto.Unmarshal(d.Password, pw)
}
