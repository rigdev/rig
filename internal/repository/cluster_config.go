package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/gen/go/capsule"
	"github.com/rigdev/rig/pkg/iterator"
)

type ClusterConfig interface {
	GetCapsuleConfig(ctx context.Context, capsuleName string) (*capsule.Config, error)
	CreateCapsuleConfig(ctx context.Context, p *capsule.Config) error
	UpdateCapsuleConfig(ctx context.Context, p *capsule.Config) error
	ListCapsuleConfigs(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.Config], int64, error)
	DeleteCapsuleConfig(ctx context.Context, capsuleName string) error

	BuildIndexes(ctx context.Context) error
}
