package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/pkg/repository/session/mongo/schema"
	"go.mongodb.org/mongo-driver/bson"
)

// Update udates when tokens metadata
func (r *MongoRepository) Update(ctx context.Context, userID, sessionID uuid.UUID, p *user.Session) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	s, err := schema.SessionFromProto(projectID, userID, sessionID, p)
	if err != nil {
		return err
	}

	res, err := r.Collection.UpdateOne(
		ctx,
		schema.GetSessionIDFilter(projectID, userID, sessionID),
		bson.M{
			"$set": s,
		},
	)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.NotFoundErrorf("session not found")
	}

	return nil
}
