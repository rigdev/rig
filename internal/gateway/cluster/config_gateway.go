package cluster

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/iterator"
)

type ConfigGateway interface {
	// GetProject(ctx context.Context, projectName string) (*project.Config, error)
	// UpsertProject(ctx context.Context, cfg *project.Config) error
	// DeleteProject(ctx context.Context, projectName string) error

	GetCapsuleConfig(ctx context.Context, capsuleName string) (*v1alpha1.Capsule, error)
	CreateCapsuleConfig(ctx context.Context, cfg *v1alpha1.Capsule) error
	UpdateCapsuleConfig(ctx context.Context, cfg *v1alpha1.Capsule) error
	ListCapsuleConfigs(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*v1alpha1.Capsule], int64, error)
	DeleteCapsuleConfig(ctx context.Context, capsuleName string) error

	SetEnvironmentVariables(ctx context.Context, capsuleName string, envs map[string]string) error
	GetEnvironmentVariables(ctx context.Context, capsuleName string) (map[string]string, error)
	SetEnvironmentVariable(ctx context.Context, capsuleName, name, value string) error
	GetEnvironmentVariable(ctx context.Context, capsuleName, name string) (value string, ok bool, err error)
	DeleteEnvironmentVariable(ctx context.Context, capsuleName, name string) error
}
