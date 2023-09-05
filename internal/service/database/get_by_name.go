package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
)

func (s *Service) GetByName(ctx context.Context, name string) (*database.Database, error) {
	return s.dr.GetByName(ctx, name)
}
