package cluster

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/iterator"
	v1 "k8s.io/api/core/v1"
)

type ConfigGateway interface {
	// GetProject(ctx context.Context, projectName string) (*project.Config, error)
	// UpsertProject(ctx context.Context, cfg *project.Config) error
	// DeleteProject(ctx context.Context, projectName string) error

	GetCapsuleConfig(ctx context.Context, capsuleID string) (*v1alpha1.Capsule, error)
	CreateCapsuleConfig(ctx context.Context, cfg *v1alpha1.Capsule) error
	UpdateCapsuleConfig(ctx context.Context, cfg *v1alpha1.Capsule) error
	ListCapsuleConfigs(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*v1alpha1.Capsule], int64, error)
	DeleteCapsuleConfig(ctx context.Context, capsuleID string) error

	SetEnvironmentVariables(ctx context.Context, capsuleID string, envs map[string]string) error
	GetEnvironmentVariables(ctx context.Context, capsuleID string) (map[string]string, error)
	SetEnvironmentVariable(ctx context.Context, capsuleID, name, value string) error
	GetEnvironmentVariable(ctx context.Context, capsuleID, name string) (value string, ok bool, err error)
	DeleteEnvironmentVariable(ctx context.Context, capsuleID, name string) error

	GetFile(ctx context.Context, capsuleID, name, namespace string) (*v1.ConfigMap, error)
	SetFile(ctx context.Context, capsuleID string, file *v1.ConfigMap) error
	ListFiles(ctx context.Context, capsuleID string, pagination *model.Pagination) (iterator.Iterator[*v1.ConfigMap], int64, error)
	DeleteFile(ctx context.Context, capsuleID, name, namespace string) error

	GetSecret(ctx context.Context, capsuleID, name, namespace string) (*v1.Secret, error)
	SetSecret(ctx context.Context, capsuleID string, file *v1.Secret) error
	ListSecrets(ctx context.Context, capsuleID string, pagination *model.Pagination) (iterator.Iterator[*v1.Secret], int64, error)
	DeleteSecret(ctx context.Context, capsuleID, name, namespace string) error
}
