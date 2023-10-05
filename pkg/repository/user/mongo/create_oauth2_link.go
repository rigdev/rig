package mongo

import (
	"context"

	"github.com/rigdev/rig/gen/go/oauth2"
	"github.com/rigdev/rig/pkg/repository/user/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

func (r *MongoRepository) CreateOauth2Link(ctx context.Context, userID uuid.UUID, p *oauth2.ProviderLink) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	v, err := schema.Oauth2LinkFromProto(projectID, userID, p)
	if err != nil {
		return err
	}

	if _, err := r.Oauth2Col.InsertOne(ctx, v); mongo.IsDuplicateKeyError(err) {
		return errors.AlreadyExistsErrorf("oauth for issuer already exists")
	} else if err != nil {
		return err
	}
	// set new data for user created
	return nil
}
