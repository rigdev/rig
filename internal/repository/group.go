package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

type Group interface {
	Create(ctx context.Context, group *group.Group) (*group.Group, error)
	Delete(ctx context.Context, groupID uuid.UUID) error
	DeleteAll(ctx context.Context) error
	Get(ctx context.Context, groupID uuid.UUID) (*group.Group, error)
	GetByName(ctx context.Context, groupName string) (*group.Group, error)
	List(ctx context.Context, pagination *model.Pagination, search string) (iterator.Iterator[*group.Group], uint64, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, group *group.Group) (*group.Group, error)
	RemoveMember(ctx context.Context, userID, groupID uuid.UUID) error
	RemoveMemberFromAll(ctx context.Context, userID uuid.UUID) error
	AddMembers(ctx context.Context, userIDs []uuid.UUID, groupID uuid.UUID) error
	ListGroupsForUser(ctx context.Context, userID uuid.UUID, pagination *model.Pagination) (iterator.Iterator[uuid.UUID], uint64, error)
	ListMembers(ctx context.Context, groupID uuid.UUID, pagination *model.Pagination) (iterator.Iterator[uuid.UUID], uint64, error)
	CountMembers(ctx context.Context, groupID uuid.UUID) (uint64, error)

	BuildIndexes(ctx context.Context) error
}
