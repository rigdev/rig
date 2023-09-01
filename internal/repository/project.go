package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

type Project interface {
	Create(ctx context.Context, p *project.Project) (*project.Project, error)
	Get(ctx context.Context, projectID uuid.UUID) (*project.Project, error)
	List(ctx context.Context, pagination *model.Pagination, exclude []uuid.UUID) (iterator.Iterator[*project.Project], int64, error)
	Update(ctx context.Context, p *project.Project) (*project.Project, error)
	Delete(ctx context.Context, projectID uuid.UUID) error
	SetSettings(ctx context.Context, projectID uuid.UUID, settingsName string, data []byte) error
	GetSettings(ctx context.Context, projectID uuid.UUID, settingsName string) ([]byte, error)
	BuildIndexes(ctx context.Context) error
}
