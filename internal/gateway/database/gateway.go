package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
)

type Gateway interface {
	Test(ctx context.Context) error
	Create(ctx context.Context, dbName string) error
	Delete(ctx context.Context, dbName string) error

	CreateTable(ctx context.Context, dbName, tableName string) error
	ListTables(ctx context.Context, dbName string) ([]*database.Table, error)
	DeleteTable(ctx context.Context, dbName, tableName string) error

	CreateCredentials(ctx context.Context, dbName, clientID, clientSecret string) error
	DeleteCredentials(ctx context.Context, dbName, clientID string) error
}
