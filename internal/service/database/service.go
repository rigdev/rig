package database

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/internal/service/user"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service struct {
	cfg config.Config
	dr  repository.Database

	mongo    *mongo.Client
	postgres *sql.DB

	logger *zap.Logger
}

type newServiceParams struct {
	fx.In

	Config       config.Config
	DatabaseRepo repository.Database
	UserService  user.Service
	Logger       *zap.Logger
	Mongo        *mongo.Client `optional:"true"`
	Postgres     *sql.DB       `optional:"true"`
}

func NewService(p newServiceParams) (*Service, error) {
	return &Service{
		cfg:      p.Config,
		dr:       p.DatabaseRepo,
		logger:   p.Logger,
		mongo:    p.Mongo,
		postgres: p.Postgres,
	}, nil
}

func applyUpdates(d *database.Database, ds []*database.Update) error {
	for _, up := range ds {
		switch v := up.GetField().(type) {
		case *database.Update_Name:
			d.Name = v.Name
		default:
			return errors.InvalidArgumentErrorf("invalid database update type '%v'", reflect.TypeOf(up.GetField()))
		}
	}
	return nil
}

func formatClientID(certificateID uuid.UUID) string {
	return fmt.Sprint("rig_", certificateID)
}

func formatDatabaseID(databaseID uuid.UUID) string {
	return fmt.Sprint("rigdb_", databaseID)
}

func (s *Service) dropMongoUser(clientID string, databaseID uuid.UUID) error {
	err := s.mongo.Database("admin").RunCommand(context.Background(), bson.D{{Key: "dropUser", Value: clientID}}).Err()
	if err != nil && strings.Contains(err.Error(), "UserNotFound") {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func (s *Service) dropPostgresUser(clientID string, databaseID uuid.UUID) error {
	if _, err := s.postgres.Exec(fmt.Sprintf("revoke all privileges on database %s from %s", formatDatabaseID(databaseID), clientID)); err != nil {
		return err
	}
	if _, err := s.postgres.Exec(fmt.Sprintf("drop user %s", clientID)); err != nil {
		return err
	}
	return nil
}

func (s *Service) mongoEnabled() error {
	if s.mongo == nil {
		return errors.FailedPreconditionErrorf("cannot register a MongoDB database without MongoDB being enabled")
	}
	return nil
}

func (s *Service) postgresEnabled() error {
	if s.postgres == nil {
		return errors.FailedPreconditionErrorf("cannot register a Postgres database without Postgres being enabled")
	}
	return nil
}
