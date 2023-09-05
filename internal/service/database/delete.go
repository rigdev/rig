package database

import (
	"context"

	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) Delete(ctx context.Context, databaseID uuid.UUID) error {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return err
	}

	gateway, err := s.getDatabaseGateway(ctx, db)
	if err != nil {
		return err
	}

	dbName := formatDatabaseID(databaseID.String())
	if err := gateway.Delete(ctx, dbName); err != nil {
		return err
	}

	if err := s.dr.Delete(ctx, databaseID); err != nil {
		return err
	}
	return nil
}
