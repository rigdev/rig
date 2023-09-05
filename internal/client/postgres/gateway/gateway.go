package postgres_gateway

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/uptrace/bun"
)

type PostgresGateway struct {
	client *bun.DB
}

func New(client *bun.DB) *PostgresGateway {
	return &PostgresGateway{
		client: client,
	}
}

func (p *PostgresGateway) Test(ctx context.Context) error {
	return p.client.Ping()
}

func (p *PostgresGateway) Create(ctx context.Context, dbName string) error {
	return errors.UnimplementedErrorf("postgres gateway unimplemented")
}

func (p *PostgresGateway) CreateTable(ctx context.Context, dbName, tableName string) error {
	return errors.UnimplementedErrorf("postgres gateway create table unimplemented")
}

func (p *PostgresGateway) ListTables(ctx context.Context, dbName string) ([]*database.Table, error) {
	return nil, errors.UnimplementedErrorf("postgres gateway list tables unimplemented")
}

func (p *PostgresGateway) DeleteTable(ctx context.Context, dbName, tableName string) error {
	return errors.UnimplementedErrorf("postgres gateway delete table unimplemented")
}

func (p *PostgresGateway) Delete(ctx context.Context, dbName string) error {
	return errors.UnimplementedErrorf("postgres gateway delete unimplemented")
}

func (p *PostgresGateway) CreateCredentials(ctx context.Context, dbName, clientID, clientSecret string) error {
	return errors.UnimplementedErrorf("postgres gateway create credentials unimplemented")
}

func (p *PostgresGateway) DeleteCredentials(ctx context.Context, dbName, clientID string) error {
	return errors.UnimplementedErrorf("postgres gateway delete credentials unimplemented")
}
