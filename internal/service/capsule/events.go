package capsule

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) CreateEvent(ctx context.Context, capsuleID uuid.UUID, rolloutID uint64, message string, ed *capsule.EventData) error {
	e := &capsule.Event{
		CreatedAt: timestamppb.Now(),
		RolloutId: rolloutID,
		Message:   message,
		EventData: ed,
	}
	if a, err := s.as.GetAuthor(ctx); err != nil {
		return err
	} else {
		e.CreatedBy = a
	}
	return s.cr.CreateEvent(ctx, capsuleID, e)
}

func (s *Service) ListEvents(ctx context.Context, capsuleID uuid.UUID, rolloutID uint64, pagination *model.Pagination) (iterator.Iterator[*capsule.Event], uint64, error) {
	return s.cr.ListEvents(ctx, pagination, capsuleID, rolloutID)
}
