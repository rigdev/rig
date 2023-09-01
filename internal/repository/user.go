package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/gen/go/oauth2"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

type User interface {
	Create(ctx context.Context, user *user.User) (*user.User, error)
	Update(ctx context.Context, update *user.User) (*user.User, error)

	GetPassword(ctx context.Context, userID uuid.UUID) (*model.HashingInstance, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, password *model.HashingInstance) error

	Get(ctx context.Context, userID uuid.UUID) (*user.User, error)
	GetByIdentifier(ctx context.Context, id *model.UserIdentifier) (*user.User, error)
	List(ctx context.Context, set *settings.Settings, pagination *model.Pagination, search string) (iterator.Iterator[*model.UserEntry], uint64, error)
	Delete(ctx context.Context, userID uuid.UUID) (*user.User, error)
	DeleteMany(ctx context.Context, userBatch []*model.UserIdentifier) (uint64, error)
	DeleteAll(ctx context.Context) error

	GetOauth2Link(ctx context.Context, issuer string, subject string) (uuid.UUID, *oauth2.ProviderLink, error)
	CreateOauth2Link(ctx context.Context, userID uuid.UUID, p *oauth2.ProviderLink) error
	DeleteOauth2Links(ctx context.Context, userId uuid.UUID) error

	BuildIndexes(ctx context.Context) error
}
