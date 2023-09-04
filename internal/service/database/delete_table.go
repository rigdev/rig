package database

import (
	"context"

	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) DeleteTable(ctx context.Context, databaseID uuid.UUID, tableName string) error {
	db, _, err := s.Get(ctx, databaseID)
	if err != nil {
		return err
	}

	gateway, err := s.getDatabaseGateway(ctx, db)
	if err != nil {
		return err
	}

	if err := gateway.DeleteTable(ctx, db.GetName(), tableName); err != nil {
		return err
	}

	for i, table := range db.Tables {
		if table.Name == tableName {
			db.Tables = append(db.Tables[:i], db.Tables[i+1:]...)
			break
		}
	}

	if _, err := s.dr.Update(ctx, db); err != nil {
		return err
	}

	return nil
}
