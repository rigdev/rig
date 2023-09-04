package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/database"
	mongo_gateway "github.com/rigdev/rig/internal/client/mongo/gateway"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func New(user, password, host string, logger *zap.Logger) (*mongo.Client, error) {
	ctx := context.Background()
	logger.Debug("trying to create mongo client...")
	withRetry := 3
	mongoUri := fmt.Sprintf("mongodb://%s:%s@%s/?retryWrites=true&w=majority", user, password, host)
	if user == "" {
		mongoUri = fmt.Sprintf("mongodb://%s/?retryWrites=true&w=majority", host)
	}
	var client *mongo.Client
	if err := utils.Retry(withRetry, time.Second*5, func() (err error) {
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(
			mongoUri,
		))
		if err != nil {
			logger.Sugar().Errorf("could not connect to MongoDB with err: %v", err)
			return err
		}
		return nil
	}); err != nil {
		logger.Sugar().Errorf("could not connect to mongo with uri %s", mongoUri)
		return nil, err
	}
	if err := PerformMongHealthCheck(ctx, client); err != nil {
		return nil, err
	}
	logger.Debug("mongo client created...")
	return client, nil
}

func NewDefault(cfg config.Config, logger *zap.Logger) (*mongo.Client, error) {
	return New(cfg.Client.Mongo.User, cfg.Client.Mongo.Password, cfg.Client.Mongo.Host, logger)
}

func NewGateway(db *database.Database, logger *zap.Logger) (*mongo_gateway.MongoGateway, error) {
	m := db.GetConfig().GetMongo()
	c, err := New(m.GetCredentials().GetPublicKey(), m.GetCredentials().GetPrivateKey(), m.GetHost(), logger)
	if err != nil {
		return nil, err
	}
	return mongo_gateway.New(c), nil
}

func PerformMongHealthCheck(ctx context.Context, client *mongo.Client) error {
	if client == nil {
		return errors.New("mongo client is nil")
	}
	if err := checkMongoConnection(ctx, client); err != nil {
		return err
	}
	return nil
}

func checkMongoConnection(ctx context.Context, client *mongo.Client) error {
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}
	return nil
}
