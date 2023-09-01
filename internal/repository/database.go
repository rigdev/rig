package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

type Database interface {
	Create(ctx context.Context, database *database.Database) (*database.Database, error)
	Delete(ctx context.Context, databaseID uuid.UUID) error
	GetByName(ctx context.Context, name string) (*database.Database, error)

	Get(ctx context.Context, databaseID uuid.UUID) (*database.Database, error)
	List(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*database.Database], uint64, error)

	Update(ctx context.Context, database *database.Database) (*database.Database, error)

	BuildIndexes(ctx context.Context) error
}
