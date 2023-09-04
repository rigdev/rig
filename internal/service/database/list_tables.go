package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) ListTables(ctx context.Context, databaseID uuid.UUID) ([]*database.Table, error) {
	db, _, err := s.Get(ctx, databaseID)
	if err != nil {
		return nil, err
	}
	return db.GetTables(), nil
}
