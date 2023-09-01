package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/internal/repository/service_account/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

func (r *MongoRepository) Create(ctx context.Context, serviceAccountID uuid.UUID, sa *service_account.ServiceAccount) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	f, err := schema.ServiceAccountFromProto(projectID, serviceAccountID, sa)
	if err != nil {
		return err
	}

	if _, err := r.Collection.InsertOne(ctx, f); mongo.IsDuplicateKeyError(err) {
		return errors.AlreadyExistsErrorf("service_account '%v' already exists", sa.GetName())
	} else if err != nil {
		return err
	}

	return nil
}
