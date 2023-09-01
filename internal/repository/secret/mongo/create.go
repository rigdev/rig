package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/internal/repository/secret/mongo/model"
	"go.mongodb.org/mongo-driver/mongo"
)

func (r *MongoRepository) Create(ctx context.Context, secretID uuid.UUID, secret []byte) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	if secret, err = r.encryptSecret(secret); err != nil {
		return err
	}

	s := model.Secret{
		ProjectID: projectID,
		SecretID:  secretID,
		Secret:    secret,
	}

	if _, err = r.Collection.InsertOne(ctx, s); mongo.IsDuplicateKeyError(err) {
		return errors.AlreadyExistsErrorf("secret already exists")
	} else if err != nil {
		return err
	}

	return nil
}
