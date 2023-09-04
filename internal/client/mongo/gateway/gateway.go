package mongo_gateway

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoGateway struct {
	client *mongo.Client
}

func New(client *mongo.Client) *MongoGateway {
	return &MongoGateway{
		client: client,
	}
}

func (m *MongoGateway) Test(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

func (m *MongoGateway) CreateTable(ctx context.Context, dbName, tableName string) error {
	return nil
}

func (m *MongoGateway) ListTables(ctx context.Context, dbName string) ([]*database.Table, error) {
	return nil, nil
}

func (m *MongoGateway) DeleteTable(ctx context.Context, dbName, tableName string) error {
	return nil
}

func (m *MongoGateway) Delete(ctx context.Context, dbName string) error {
	return nil
}
