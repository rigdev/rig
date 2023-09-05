package database

import (
	"context"
	"fmt"

	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) CreateCredentials(ctx context.Context, clientId string, databaseID uuid.UUID) (clientSecret string, err error) {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return "", err
	}
	if clientId == "" {
		return "", errors.InvalidArgumentErrorf("credential name cannot be empty")
	}

	secret := uuid.New()
	clientSecret = fmt.Sprint("secret_", secret.String())

	gateway, err := s.getDatabaseGateway(ctx, db)
	if err != nil {
		return "", err
	}

	dbName := formatDatabaseID(databaseID.String())
	if err := gateway.CreateCredentials(ctx, dbName, clientId, clientSecret); err != nil {
		return "", err
	}

	db.ClientIds = append(db.ClientIds, clientId)
	if _, err := s.dr.Update(ctx, db); err != nil {
		return "", err
	}

	return clientSecret, nil
}
