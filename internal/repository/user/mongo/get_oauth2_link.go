package mongo

import (
	"context"

	"github.com/rigdev/rig/gen/go/oauth2"
	"github.com/rigdev/rig/internal/repository/user/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (m *MongoRepository) GetOauth2Link(ctx context.Context, issuer string, subject string) (uuid.UUID, *oauth2.ProviderLink, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return uuid.Nil, nil, err
	}
	filter := schema.GetOauth2UserFilter(projectID, issuer, subject)
	v := schema.Oauth2Link{}
	if err := m.Oauth2Col.FindOne(ctx, filter).Decode(&v); err != nil {
		return uuid.Nil, nil, errors.NotFoundErrorf("oauth2 link not found")
	}

	p, err := v.ToProto()
	if err != nil {
		return uuid.Nil, nil, err
	}

	return v.UserID, p, nil
}
