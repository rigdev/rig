package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func New(cfg config.Config, logger *zap.Logger) (*mongo.Client, error) {
	ctx := context.Background()
	logger.Debug("trying to create mongo client...")
	withRetry := 3
	mongoUri := fmt.Sprintf("mongodb://%s:%s@%s/?retryWrites=true&w=majority", cfg.Client.Mongo.User, cfg.Client.Mongo.Password, cfg.Client.Mongo.Host)
	if cfg.Client.Mongo.User == "" {
		mongoUri = fmt.Sprintf("mongodb://%s/?retryWrites=true&w=majority", cfg.Client.Mongo.Host)
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
