package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/pkg/repository/secret/mongo/model"
	"go.mongodb.org/mongo-driver/mongo"
)

func (r *MongoRepository) Get(ctx context.Context, secretID uuid.UUID) ([]byte, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := model.GetSecretIDFilter(projectID, secretID)
	s := model.Secret{}
	if err := r.Collection.FindOne(ctx, filter).Decode(&s); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("secret not found")
	} else if err != nil {
		return nil, err
	}

	return r.decryptSecret(s.Secret)
}
