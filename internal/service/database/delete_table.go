package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) DeleteTable(ctx context.Context, databaseID uuid.UUID, collectionName string) error {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return err
	}
	switch db.Type {
	case database.Type_TYPE_MONGO:
		if err := s.mongoEnabled(); err != nil {
			return err
		}
		if err := s.mongo.Database(formatDatabaseID(databaseID)).Collection(collectionName).Drop(ctx); err != nil {
			return err
		}
	default:
		return errors.InternalErrorf("invalid database type: %v", db.Type)
	}
	return nil
}
