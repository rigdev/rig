package database

import (
	"context"
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) CreateTable(ctx context.Context, databaseID uuid.UUID, tableName string) error {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return err
	}
	switch db.Type {
	case database.Type_TYPE_MONGO:
		if err := s.mongoEnabled(); err != nil {
			return err
		}
		if err := s.mongo.Database(formatDatabaseID(databaseID)).CreateCollection(ctx, tableName); err != nil {
			return err
		}
	case database.Type_TYPE_POSTGRES:
		if err := s.postgresEnabled(); err != nil {
			return err
		}
		if _, err := s.postgres.Exec(fmt.Sprintf("create table %s ()", tableName)); err != nil {
			return err
		}
		// TODO: parse table formats
		// int2, int4, int8, float4, float8, numeric, json, text, varchar, uuid, date, time, timetz, timestamp, timestamptz, bool
	default:
		return errors.InternalErrorf("invalid database type: %v", db.Type)
	}
	return nil
}
