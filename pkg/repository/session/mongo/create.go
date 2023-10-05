package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/pkg/repository/session/mongo/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

func (r *MongoRepository) Create(ctx context.Context, userID, sessionID uuid.UUID, p *user.Session) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	s, err := schema.SessionFromProto(projectID, userID, sessionID, p)
	if err != nil {
		return err
	}

	if _, err = r.Collection.InsertOne(ctx, s); mongo.IsDuplicateKeyError(err) {
		return errors.AlreadyExistsErrorf("session already exists")
	} else if err != nil {
		return err
	}

	return nil
}
