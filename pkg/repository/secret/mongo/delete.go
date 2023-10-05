package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/pkg/repository/secret/mongo/model"
	"go.mongodb.org/mongo-driver/mongo"
)

func (t *MongoRepository) Delete(ctx context.Context, secretID uuid.UUID) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	filter := model.GetSecretIDFilter(projectID, secretID)
	if _, err := t.Collection.DeleteOne(ctx, filter); err == mongo.ErrNoDocuments {
		return errors.NotFoundErrorf("secret not found")
	} else if err != nil {
		return err
	}

	return nil
}
