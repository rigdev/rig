package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) DeleteCredential(ctx context.Context, credentialsName string, databaseID uuid.UUID) error {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return err
	}
	isFound := false
	for index, credential := range db.GetInfo().GetCredentials() {
		if credential.GetName() == credentialsName {
			switch db.GetType() {
			case database.Type_TYPE_MONGO:
				if err := s.mongoEnabled(); err != nil {
					return err
				}
				if err := s.dropMongoUser(credential.ClientId, databaseID); err != nil {
					return err
				}
			case database.Type_TYPE_POSTGRES:
				if err := s.postgresEnabled(); err != nil {
					return err
				}
				if err := s.dropPostgresUser(credential.ClientId, databaseID); err != nil {
					return err
				}
			default:
				return errors.InternalErrorf("invalid database type: %v", db.Type)

			}
			db.Info.Credentials = append(db.Info.Credentials[:index], db.Info.Credentials[index+1:]...)
			isFound = true
			break
		}
	}
	if !isFound {
		return errors.NotFoundErrorf("could not find clientId credential")
	}
	if _, err := s.dr.Update(ctx, db); err != nil {
		return err
	}
	return nil
}
