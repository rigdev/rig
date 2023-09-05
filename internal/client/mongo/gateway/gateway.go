package mongo_gateway

import (
	"context"
	"strings"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"go.mongodb.org/mongo-driver/bson"
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

func (m *MongoGateway) Create(ctx context.Context, dbName string) error {
	return nil
}

func (m *MongoGateway) CreateTable(ctx context.Context, dbName, tableName string) error {
	return m.client.Database(dbName).CreateCollection(ctx, tableName)
}

func (m *MongoGateway) ListTables(ctx context.Context, dbName string) ([]*database.Table, error) {
	collections, err := m.client.Database(dbName).ListCollectionNames(ctx, nil)
	if err != nil {
		return nil, err
	}
	var tables []*database.Table
	for _, c := range collections {
		tables = append(tables, &database.Table{
			Name: c,
		})
	}
	return tables, nil
}

func (m *MongoGateway) DeleteTable(ctx context.Context, dbName, tableName string) error {
	return m.client.Database(dbName).Collection(tableName).Drop(ctx)
}

func (m *MongoGateway) Delete(ctx context.Context, dbName string) error {
	return m.client.Database(dbName).Drop(ctx)
}

func (m *MongoGateway) CreateCredentials(ctx context.Context, dbName, clientID, clientSecret string) error {
	if err := m.client.Database("admin").RunCommand(context.Background(), bson.D{
		{Key: "createUser", Value: clientID},
		{Key: "pwd", Value: clientSecret},
		{Key: "roles", Value: []bson.M{{"role": "readWrite", "db": dbName}}},
	}); err.Err() != nil {
		return err.Err()
	}
	return nil
}

func (m *MongoGateway) DeleteCredentials(ctx context.Context, dbName, clientID string) error {
	err := m.client.Database("admin").RunCommand(context.Background(), bson.D{{Key: "dropUser", Value: clientID}}).Err()
	if err != nil && strings.Contains(err.Error(), "UserNotFound") {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}
