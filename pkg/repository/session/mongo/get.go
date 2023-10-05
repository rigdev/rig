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

func (t *MongoRepository) Get(ctx context.Context, userID, sessionID uuid.UUID) (*user.Session, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := schema.GetSessionIDFilter(projectID, userID, sessionID)
	s := schema.Session{}
	if err := t.Collection.FindOne(ctx, filter).Decode(&s); err == mongo.ErrNoDocuments {
		return nil, errors.NotFoundErrorf("session not found")
	} else if err != nil {
		return nil, err
	}
	return s.ToProto()
}
