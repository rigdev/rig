package database

import (
	"context"
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) Delete(ctx context.Context, databaseID uuid.UUID) error {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return err
	}
	switch db.Type {
	case database.Type_TYPE_MONGO:
		if err := s.mongoEnabled(); err != nil {
			return err
		}
		// delete all database user in that database
		for _, credential := range db.GetInfo().GetCredentials() {
			if err := s.dropMongoUser(credential.ClientId, databaseID); err != nil {
				return err
			}
		}
		if err := s.mongo.Database(formatDatabaseID(databaseID)).Drop(ctx); err != nil {
			return err
		}
	case database.Type_TYPE_POSTGRES:
		if err := s.postgresEnabled(); err != nil {
			return err
		}
		for _, credential := range db.GetInfo().GetCredentials() {
			if err := s.dropPostgresUser(credential.ClientId, databaseID); err != nil {
				return err
			}
		}
		if _, err := s.postgres.Exec(fmt.Sprintf("drop database %s", formatDatabaseID(databaseID))); err != nil {
			return err
		}
	default:
		return errors.InternalErrorf("invalid database type: %v", db.Type)
	}
	if err := s.dr.Delete(ctx, databaseID); err != nil {
		return err
	}
	return nil
}
