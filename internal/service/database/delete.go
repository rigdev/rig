package database

import (
	"context"

	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) Delete(ctx context.Context, databaseID uuid.UUID) error {
	db, sid, err := s.Get(ctx, databaseID)
	if err != nil {
		return err
	}

	gateway, err := s.getDatabaseGateway(ctx, db)
	if err != nil {
		return err
	}

	if err := gateway.Delete(ctx, db.GetName()); err != nil {
		return err
	}

	if err := s.secr.Delete(ctx, sid); err != nil {
		return err
	}

	if err := s.dr.Delete(ctx, databaseID); err != nil {
		return err
	}
	return nil
}
