package database

import (
	"context"
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/mongo"

	mongo_gateway "github.com/rigdev/rig/internal/client/mongo/gateway"
	postgres_gateway "github.com/rigdev/rig/internal/client/postgres/gateway"
	"github.com/rigdev/rig/internal/config"
	database_gateway "github.com/rigdev/rig/internal/gateway/database"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service struct {
	cfg         config.Config
	dr          repository.Database
	secr        repository.Secret
	logger      *zap.Logger
	MongoClient *mongo.Client `optional:"true"`
	Postgres    *bun.DB       `optional:"true"`
}

type newServiceParams struct {
	fx.In
	Config       config.Config
	DatabaseRepo repository.Database
	SecretRepo   repository.Secret
	Logger       *zap.Logger
	MongoClient  *mongo.Client `optional:"true"`
	Postgres     *bun.DB       `optional:"true"`
}

func NewService(p newServiceParams) (*Service, error) {
	return &Service{
		cfg:         p.Config,
		dr:          p.DatabaseRepo,
		secr:        p.SecretRepo,
		logger:      p.Logger,
		MongoClient: p.MongoClient,
		Postgres:    p.Postgres,
	}, nil
}

func (s *Service) getDatabaseGateway(ctx context.Context, db *database.Database) (database_gateway.Gateway, error) {
	switch db.GetType() {
	case database.Type_TYPE_MONGODB:
		if s.MongoClient == nil {
			return nil, errors.FailedPreconditionErrorf("no mongo client configured, RIG_CLIENT_MONGO_HOST environment variable missing")
		}
		return mongo_gateway.New(s.MongoClient), nil
	case database.Type_TYPE_POSTGRES:
		if s.Postgres == nil {
			return nil, errors.FailedPreconditionErrorf("no postgres client configured, RIG_CLIENT_POSTGRES_HOST environment variable missing")
		}
		return postgres_gateway.New(s.Postgres), nil
	default:
		return nil, errors.InvalidArgumentErrorf("invalid database type '%v'", db.GetType())
	}
}

func formatDatabaseID(databaseID string) string {
	return fmt.Sprint("rigdb_", databaseID)
}
