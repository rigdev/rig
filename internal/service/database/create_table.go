package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) CreateTable(ctx context.Context, databaseID uuid.UUID, tableName string) error {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return err
	}

	gateway, err := s.getDatabaseGateway(ctx, db)
	if err != nil {
		return err
	}

	dbName := formatDatabaseID(databaseID.String())
	if err := gateway.CreateTable(ctx, dbName, tableName); err != nil {
		return err
	}

	db.Tables = append(db.Tables, &database.Table{
		Name: tableName,
	})

	if _, err := s.dr.Update(ctx, db); err != nil {
		return err
	}

	return nil
}
