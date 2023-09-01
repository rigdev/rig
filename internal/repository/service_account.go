package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

type ServiceAccount interface {
	Create(ctx context.Context, serviceAccountID uuid.UUID, c *service_account.ServiceAccount) error
	List(ctx context.Context) (iterator.Iterator[*service_account.Entry], error)
	Get(ctx context.Context, serviceAccountID uuid.UUID) (uuid.UUID, *service_account.ServiceAccount, error)
	GetClientSecret(ctx context.Context, serviceAccountID uuid.UUID) (*model.HashingInstance, error)
	UpdateClientSecret(ctx context.Context, serviceAccountID uuid.UUID, pw *model.HashingInstance) error
	Delete(ctx context.Context, serviceAccountID uuid.UUID) error
	BuildIndexes(ctx context.Context) error
}
