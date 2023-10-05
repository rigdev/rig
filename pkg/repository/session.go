package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
)

type Session interface {
	Create(ctx context.Context, userID, sessionID uuid.UUID, s *user.Session) error
	Update(ctx context.Context, userID, sessionID uuid.UUID, s *user.Session) error
	List(ctx context.Context, userID uuid.UUID) (iterator.Iterator[*user.SessionEntry], error)
	Get(ctx context.Context, userID, sessionID uuid.UUID) (*user.Session, error)
	Delete(ctx context.Context, userID, sessionID uuid.UUID) error
	BuildIndexes(ctx context.Context) error
}
