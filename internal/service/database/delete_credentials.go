package database

import (
	"context"

	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) DeleteCredentials(ctx context.Context, clientId string, databaseID uuid.UUID) error {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return err
	}

	isFound := false
	for index, cId := range db.GetClientIds() {
		if cId == clientId {
			db.ClientIds = append(db.ClientIds[:index], db.ClientIds[index+1:]...)
			isFound = true
			break
		}
	}
	if !isFound {
		return errors.NotFoundErrorf("could not find clientId credential")
	}

	gateway, err := s.getDatabaseGateway(ctx, db)
	if err != nil {
		return err
	}

	dbName := formatDatabaseID(databaseID.String())
	if err := gateway.DeleteCredentials(ctx, dbName, clientId); err != nil {
		return err
	}

	if _, err := s.dr.Update(ctx, db); err != nil {
		return err
	}
	return nil
}
