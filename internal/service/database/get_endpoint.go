package database

import (
	"context"
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) GetEndpoint(ctx context.Context, databaseID uuid.UUID, clientID, clientSecret string) (string, string, error) {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return "", "", err
	}
	if clientID != "" {
		found := false
		for _, cID := range db.GetClientIds() {
			if cID == clientID {
				found = true
				break
			}
		}
		if !found {
			return "", "", errors.NotFoundErrorf("could not find clientID credential")
		}
	}
	if clientID == "" && clientSecret == "" {
		clientID = "username"
		clientSecret = "password"
	}
	dbName := formatDatabaseID(databaseID.String())
	switch db.GetType() {
	case database.Type_TYPE_MONGODB:
		return fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=admin", clientID, clientSecret, s.cfg.Client.Mongo.Host, dbName), dbName, nil
	case database.Type_TYPE_POSTGRES:
		return "", "", errors.UnimplementedErrorf("not currently available for postgres")
	default:
		return "", "", errors.InternalErrorf("invalid database type: %v", db.GetType())
	}
}
