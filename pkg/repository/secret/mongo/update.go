package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/pkg/repository/secret/mongo/model"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *MongoRepository) Update(ctx context.Context, secretID uuid.UUID, secret []byte) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	if secret, err = r.encryptSecret(secret); err != nil {
		return err
	}

	res, err := r.Collection.UpdateOne(
		ctx,
		model.GetSecretIDFilter(projectID, secretID),
		bson.M{
			"$set": bson.M{
				"secret": secret,
			},
		},
	)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.NotFoundErrorf("secret not found")
	}

	return nil
}
