package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/gen/go/rollout"
	repo_capsule "github.com/rigdev/rig/internal/repository/capsule"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

type Capsule interface {
	Create(ctx context.Context, c *capsule.Capsule) error
	Get(ctx context.Context, capsuleID uuid.UUID) (*capsule.Capsule, error)
	GetByName(ctx context.Context, name string) (*capsule.Capsule, error)
	List(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.Capsule], int64, error)
	Update(ctx context.Context, p *capsule.Capsule) error
	Delete(ctx context.Context, capsuleID uuid.UUID) error

	CreateBuild(ctx context.Context, capsuleID uuid.UUID, b *capsule.Build) error
	DeleteBuild(ctx context.Context, capsuleID uuid.UUID, buildID string) error
	ListBuilds(ctx context.Context, pagination *model.Pagination, capsuleID uuid.UUID) (iterator.Iterator[*capsule.Build], uint64, error)
	GetBuild(ctx context.Context, capsuleID uuid.UUID, buildID string) (*capsule.Build, error)

	CreateRollout(ctx context.Context, capsuleID uuid.UUID, rc *capsule.RolloutConfig, rs *rollout.Status) (uint64, error)
	ListRollouts(ctx context.Context, pagination *model.Pagination, capsuleID uuid.UUID) (iterator.Iterator[*capsule.Rollout], uint64, error)
	UpdateRolloutStatus(ctx context.Context, capsuleID uuid.UUID, rolloutID uint64, version uint64, rs *rollout.Status) error
	GetRollout(ctx context.Context, capsuleID uuid.UUID, rolloutID uint64) (*capsule.RolloutConfig, *rollout.Status, uint64, error)
	ActiveRollouts(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[repo_capsule.ActiveRollout], error)

	CreateEvent(ctx context.Context, capsuleID uuid.UUID, e *capsule.Event) error
	ListEvents(ctx context.Context, pagination *model.Pagination, capsuleID uuid.UUID, rolloutID uint64) (iterator.Iterator[*capsule.Event], uint64, error)

	CreateMetrics(ctx context.Context, metrics *capsule.InstanceMetrics) error
	ListMetrics(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.InstanceMetrics], error)
	GetMetrics(ctx context.Context, pagination *model.Pagination, capsuleID uuid.UUID) (iterator.Iterator[*capsule.InstanceMetrics], error)
	GetInstanceMetrics(ctx context.Context, pagination *model.Pagination, capsuleID uuid.UUID, instanceID string) (iterator.Iterator[*capsule.InstanceMetrics], error)

	BuildIndexes(ctx context.Context) error
}
