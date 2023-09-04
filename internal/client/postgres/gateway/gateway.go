package postgres_gateway

import (
	"context"
	"database/sql"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/uptrace/bun"
)

type PostgresGateway struct {
	sqlClient *sql.DB
	client    *bun.DB
}

func New(sqlClient *sql.DB, client *bun.DB) *PostgresGateway {
	return &PostgresGateway{
		sqlClient: sqlClient,
		client:    client,
	}
}

func (p *PostgresGateway) Test(ctx context.Context) error {
	return p.sqlClient.Ping()
}

func (p *PostgresGateway) CreateTable(ctx context.Context, dbName, tableName string) error {
	return nil
}

func (p *PostgresGateway) ListTables(ctx context.Context, dbName string) ([]*database.Table, error) {
	return nil, nil
}

func (p *PostgresGateway) DeleteTable(ctx context.Context, dbName, tableName string) error {
	return nil
}

func (p *PostgresGateway) Delete(ctx context.Context, dbName string) error {
	return nil
}
