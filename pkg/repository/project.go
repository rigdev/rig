package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/iterator"
)

type Project interface {
	Create(ctx context.Context, p *project.Project) (*project.Project, error)
	Get(ctx context.Context, projectID string) (*project.Project, error)
	List(ctx context.Context, pagination *model.Pagination, exclude []string) (iterator.Iterator[*project.Project], int64, error)
	Update(ctx context.Context, p *project.Project) (*project.Project, error)
	Delete(ctx context.Context, projectID string) error
	SetSettings(ctx context.Context, projectID string, settingsName string, data []byte) error
	GetSettings(ctx context.Context, projectID string, settingsName string) ([]byte, error)
	BuildIndexes(ctx context.Context) error
}
