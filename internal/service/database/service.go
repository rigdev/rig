package database

import (
	"context"
	"reflect"

	"github.com/rigdev/rig-go-api/api/v1/database"

	"github.com/rigdev/rig/internal/client/mongo"
	"github.com/rigdev/rig/internal/client/postgres"
	"github.com/rigdev/rig/internal/config"
	database_gateway "github.com/rigdev/rig/internal/gateway/database"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service struct {
	cfg  config.Config
	dr   repository.Database
	secr repository.Secret

	logger *zap.Logger
}

type newServiceParams struct {
	fx.In
	Config       config.Config
	DatabaseRepo repository.Database
	SecretRepo   repository.Secret
	Logger       *zap.Logger
}

func NewService(p newServiceParams) (*Service, error) {
	return &Service{
		cfg:    p.Config,
		dr:     p.DatabaseRepo,
		secr:   p.SecretRepo,
		logger: p.Logger,
	}, nil
}

func (s *Service) getDatabaseGateway(ctx context.Context, db *database.Database) (database_gateway.Gateway, error) {
	switch v := db.GetConfig().GetConfig().(type) {
	case *database.Config_Mongo:
		return mongo.NewGateway(db, s.logger)
	case *database.Config_Postgres:
		return postgres.NewGateway(db, s.logger)
	default:
		return nil, errors.InvalidArgumentErrorf("invalid database type '%v'", reflect.TypeOf(v))
	}
}
