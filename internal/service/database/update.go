package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) Update(ctx context.Context, databaseID uuid.UUID, updates []*database.Update) (*database.Database, error) {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return nil, err
	}
	if err := applyUpdates(db, updates); err != nil {
		return nil, err
	}
	db, err = s.dr.Update(ctx, db)
	if err != nil {
		return nil, err
	}
	return db, nil
}
