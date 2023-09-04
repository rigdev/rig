package database

import (
	"context"
	"encoding/json"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) Create(ctx context.Context, name string, config *database.Config, linkTables bool) (*database.Database, error) {
	if name == "" {
		return nil, errors.InvalidArgumentErrorf("missing required database name")
	}

	databaseID := uuid.New()
	d := &database.Database{
		DatabaseId: databaseID.String(),
		Name:       name,
		Config:     config,
		Tables:     []*database.Table{},
		CreatedAt:  timestamppb.Now(),
	}

	gateway, err := s.getDatabaseGateway(ctx, d)
	if err != nil {
		return nil, err
	}

	err = gateway.Test(ctx)
	if err != nil {
		return nil, err
	}

	sID := uuid.New()
	var secret []byte

	switch config.Config.(type) {
	case *database.Config_Mongo:
		secret, err = json.Marshal(config.GetMongo().GetCredentials())
		if err != nil {
			return nil, err
		}

		config.GetMongo().Credentials = nil
	case *database.Config_Postgres:
		secret, err = json.Marshal(config.GetPostgres().GetCredentials())
		if err != nil {
			return nil, err
		}

		config.GetPostgres().Credentials = nil
	}

	err = s.secr.Create(ctx, sID, secret)
	if err != nil {
		return nil, err
	}

	d, err = s.dr.Create(ctx, sID, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}
