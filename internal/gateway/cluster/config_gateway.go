package cluster

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/gen/go/capsule"
	"github.com/rigdev/rig/pkg/iterator"
)

type ConfigGateway interface {
	// GetProject(ctx context.Context, projectName string) (*project.Config, error)
	// UpsertProject(ctx context.Context, cfg *project.Config) error
	// DeleteProject(ctx context.Context, projectName string) error

	GetCapsuleConfig(ctx context.Context, capsuleName string) (*capsule.Config, error)
	CreateCapsuleConfig(ctx context.Context, cfg *capsule.Config) error
	UpdateCapsuleConfig(ctx context.Context, cfg *capsule.Config) error
	ListCapsuleConfigs(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.Config], int64, error)
	DeleteCapsuleConfig(ctx context.Context, capsuleName string) error
}
