package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) Create(ctx context.Context, dbType database.Type, initializers []*database.Update) (uuid.UUID, *database.Database, error) {
	databaseID := uuid.New()

	switch dbType {
	case database.Type_TYPE_MONGO:
		if err := s.mongoEnabled(); err != nil {
			return uuid.Nil, nil, err
		}
	case database.Type_TYPE_POSTGRES:
		if err := s.postgresEnabled(); err != nil {
			return uuid.Nil, nil, err
		}
		if _, err := s.postgres.Exec("create database " + formatDatabaseID(databaseID)); err != nil {
			return uuid.Nil, nil, err
		}
	default:
		return uuid.Nil, nil, errors.InvalidArgumentErrorf("invalid database type: %v", dbType)
	}

	d := &database.Database{
		DatabaseId: databaseID.String(),
		Type:       dbType,
		Info: &database.Info{
			CreatedAt:   timestamppb.Now(),
			Credentials: []*database.Credential{},
		},
	}
	if err := applyUpdates(d, initializers); err != nil {
		return uuid.Nil, nil, err
	}
	if d.Name == "" {
		return uuid.Nil, nil, errors.InvalidArgumentErrorf("missing required database name")
	}
	var err error
	d, err = s.dr.Create(ctx, d)
	if err != nil {
		return uuid.Nil, nil, err
	}
	return databaseID, d, nil
}
