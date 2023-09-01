package auth

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID) (iterator.Iterator[*user.SessionEntry], error) {
	return s.sr.List(ctx, userID)
}

func (s *Service) newSession(ctx context.Context, userID uuid.UUID, am *user.AuthMethod) (uuid.UUID, *user.Session, error) {
	sessionID := uuid.New()
	ss := &user.Session{
		CreatedAt:  timestamppb.Now(),
		AuthMethod: am,
	}

	if err := s.sr.Create(ctx, userID, sessionID, ss); err != nil {
		return uuid.Nil, nil, err
	}

	return sessionID, ss, nil
}

func (s *Service) deleteSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	return s.sr.Delete(ctx, userID, sessionID)
}
