package database

import (
	"context"
	"fmt"

	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) CreateCredentials(ctx context.Context, databaseID uuid.UUID) (string, string, error) {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return "", "", err
	}

	clientId := fmt.Sprint("rig_", uuid.New())
	clientSecret := fmt.Sprint("secret_", uuid.New())

	gateway, err := s.getDatabaseGateway(ctx, db)
	if err != nil {
		return "", "", err
	}

	dbName := formatDatabaseID(databaseID.String())
	if err := gateway.CreateCredentials(ctx, dbName, clientId, clientSecret); err != nil {
		return "", "", err
	}

	db.ClientIds = append(db.ClientIds, clientId)
	if _, err := s.dr.Update(ctx, db); err != nil {
		return "", "", err
	}

	return clientId, clientSecret, nil
}
