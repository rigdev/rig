package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) Create(ctx context.Context, name string, dbType database.Type) (string, *database.Database, error) {
	if name == "" {
		return "", nil, errors.InvalidArgumentErrorf("missing required database name")
	}

	databaseID := uuid.New()
	d := &database.Database{
		DatabaseId: databaseID.String(),
		Name:       name,
		Type:       dbType,
		Tables:     []*database.Table{},
		CreatedAt:  timestamppb.Now(),
	}

	gateway, err := s.getDatabaseGateway(ctx, d)
	if err != nil {
		return "", nil, err
	}

	dbName := formatDatabaseID(databaseID.String())
	err = gateway.Create(ctx, dbName)
	if err != nil {
		return "", nil, err
	}

	clientSecret, err := s.CreateCredentials(ctx, "default", databaseID)
	if err != nil {
		return "", nil, err
	}

	d, err = s.dr.Create(ctx, d)
	if err != nil {
		return "", nil, err
	}
	return clientSecret, d, nil
}
