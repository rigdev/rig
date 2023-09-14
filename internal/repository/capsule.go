package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/gen/go/rollout"
	repo_capsule "github.com/rigdev/rig/internal/repository/capsule"
	"github.com/rigdev/rig/pkg/iterator"
)

type Capsule interface {
	CreateBuild(ctx context.Context, capsuleID string, b *capsule.Build) error
	DeleteBuild(ctx context.Context, capsuleID string, buildID string) error
	ListBuilds(ctx context.Context, pagination *model.Pagination, capsuleID string) (iterator.Iterator[*capsule.Build], uint64, error)
	GetBuild(ctx context.Context, capsuleID string, buildID string) (*capsule.Build, error)

	CreateRollout(ctx context.Context, capsuleID string, rc *capsule.RolloutConfig, rs *rollout.Status) (uint64, error)
	ListRollouts(ctx context.Context, pagination *model.Pagination, capsuleID string) (iterator.Iterator[*capsule.Rollout], uint64, error)
	UpdateRolloutStatus(ctx context.Context, capsuleID string, rolloutID uint64, version uint64, rs *rollout.Status) error
	GetRollout(ctx context.Context, capsuleID string, rolloutID uint64) (*capsule.RolloutConfig, *rollout.Status, uint64, error)
	GetCurrentRollout(ctx context.Context, capsuleID string) (uint64, *capsule.RolloutConfig, *rollout.Status, uint64, error)
	ActiveRollouts(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[repo_capsule.ActiveRollout], error)

	CreateEvent(ctx context.Context, capsuleID string, e *capsule.Event) error
	ListEvents(ctx context.Context, pagination *model.Pagination, capsuleID string, rolloutID uint64) (iterator.Iterator[*capsule.Event], uint64, error)

	CreateMetrics(ctx context.Context, metrics *capsule.InstanceMetrics) error
	ListMetrics(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.InstanceMetrics], error)
	GetMetrics(ctx context.Context, pagination *model.Pagination, capsuleID string) (iterator.Iterator[*capsule.InstanceMetrics], error)
	GetInstanceMetrics(ctx context.Context, pagination *model.Pagination, capsuleID string, instanceID string) (iterator.Iterator[*capsule.InstanceMetrics], error)

	Delete(ctx context.Context, capsuleID string) error

	BuildIndexes(ctx context.Context) error
}
