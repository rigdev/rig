package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) GetByName(ctx context.Context, name string) (*database.Database, uuid.UUID, error) {
	return s.dr.GetByName(ctx, name)
}
