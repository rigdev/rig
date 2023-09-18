package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/iterator"
)

type ClusterConfig interface {
	GetCapsuleConfig(ctx context.Context, capsuleName string) (*v1alpha1.Capsule, error)
	CreateCapsuleConfig(ctx context.Context, p *v1alpha1.Capsule) error
	UpdateCapsuleConfig(ctx context.Context, p *v1alpha1.Capsule) error
	ListCapsuleConfigs(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*v1alpha1.Capsule], int64, error)
	DeleteCapsuleConfig(ctx context.Context, capsuleName string) error

	SetEnvironmentVariables(ctx context.Context, capsuleName string, envs map[string]string) error
	GetEnvironmentVariables(ctx context.Context, capsuleName string) (map[string]string, error)

	BuildIndexes(ctx context.Context) error
}
