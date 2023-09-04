package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
)

type Gateway interface {
	Test(ctx context.Context) error

	CreateTable(ctx context.Context, dbName, tableName string) error
	ListTables(ctx context.Context, dbName string) ([]*database.Table, error)
	DeleteTable(ctx context.Context, dbName, tableName string) error

	Delete(ctx context.Context, dbName string) error
}
