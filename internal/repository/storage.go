package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

type Storage interface {
	Create(ctx context.Context, secretID uuid.UUID, provider *storage.Provider) (*storage.Provider, error)
	Delete(ctx context.Context, providerID uuid.UUID) error
	Get(ctx context.Context, providerID uuid.UUID) (*storage.Provider, uuid.UUID, error)
	List(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*storage.Provider], uint64, error)
	Update(ctx context.Context, providerID uuid.UUID, provider *storage.Provider) (*storage.Provider, error)
	LookupByBucket(ctx context.Context, bucket string) (uuid.UUID, *storage.Provider, uuid.UUID, error)
	Lookup(ctx context.Context, name string) (uuid.UUID, *storage.Provider, uuid.UUID, error)
	BuildIndexes(ctx context.Context) error
}
