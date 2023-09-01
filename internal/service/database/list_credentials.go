package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) ListCredentials(ctx context.Context, databaseID uuid.UUID) ([]*database.Credential, error) {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return nil, err
	}
	return db.GetInfo().GetCredentials(), nil
}
