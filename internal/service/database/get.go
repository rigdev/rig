package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) Get(ctx context.Context, databaseID uuid.UUID) (*database.Database, error) {
	return s.dr.Get(ctx, databaseID)

}
