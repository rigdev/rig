package mongo

import (
	"context"
	"time"

	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rigdev/rig/pkg/repository/group/mongo/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

func (p *MongoRepository) AddMembers(ctx context.Context, userID []uuid.UUID, groupID uuid.UUID) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}
	gms := make([]interface{}, 0, len(userID))
	for _, u := range userID {
		gms = append(gms, &schema.GroupMember{
			GroupID:   groupID,
			UserID:    u,
			ProjectID: projectID,
			CreatedAt: time.Now(),
		})
	}

	if _, err := p.MembersCollection.InsertMany(ctx, gms); mongo.IsDuplicateKeyError(err) {
		return errors.AlreadyExistsErrorf("the user is already a member of the group")
	} else if err != nil {
		return err
	}

	return nil
}
