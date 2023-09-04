package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/database"
	postgres_gateway "github.com/rigdev/rig/internal/client/postgres/gateway"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.uber.org/zap"
)

func New(user, password, host string, logger *zap.Logger) (*bun.DB, *sql.DB, error) {
	ctx := context.Background()
	logger.Info("trying to create postgres client...")
	withRetry := 3
	postgresUri := fmt.Sprintf("postgres://%s:%s@%s", user, password, host)

	var client *bun.DB
	var sqldb *sql.DB
	if err := utils.Retry(withRetry, time.Second*5, func() (err error) {
		sqldb = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(postgresUri)))
		client = bun.NewDB(sqldb, pgdialect.New())
		return client.Ping()
	}); err != nil {
		logger.Sugar().Errorf(fmt.Sprintf("could not connect to postgres with uri %s", postgresUri))
		return nil, nil, err
	}
	if err := PerformPostgresHealthCheck(ctx, client); err != nil {
		return nil, nil, err
	}
	logger.Info("postgres client created...")

	return client, sqldb, nil
}

func NewDefault(cfg config.Config, logger *zap.Logger) (*bun.DB, *sql.DB, error) {
	return New(cfg.Client.Postgres.User, cfg.Client.Postgres.Password, cfg.Client.Postgres.Host, logger)
}

func NewGateway(db *database.Database, logger *zap.Logger) (*postgres_gateway.PostgresGateway, error) {
	p := db.GetConfig().GetPostgres()
	bun, sql, err := New(p.GetCredentials().GetPublicKey(), p.GetCredentials().GetPrivateKey(), p.GetHost(), logger)
	if err != nil {
		return nil, err
	}
	return postgres_gateway.New(sql, bun), nil
}

func PerformPostgresHealthCheck(ctx context.Context, client *bun.DB) error {
	if client == nil {
		return errors.New("postgres client is nil")
	}
	if err := checkPostgresConnection(ctx, client); err != nil {
		return err
	}
	return nil
}

func checkPostgresConnection(ctx context.Context, client *bun.DB) error {
	if err := client.Ping(); err != nil {
		return err
	}
	return nil
}
