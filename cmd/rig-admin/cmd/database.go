package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-api/model"
	service_database "github.com/rigdev/rig/internal/service/database"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var databaseId string

func init() {
	database := &cobra.Command{
		Use: "database",
	}
	database.PersistentFlags().StringVar(&databaseId, "database", "", "uuid of your database")

	createDatabase := &cobra.Command{
		Use:  "create-database <name> <type> <username> <password> <host>",
		RunE: register(CreateDatabase),
		Args: cobra.ExactArgs(5),
	}
	database.AddCommand(createDatabase)

	getDatabase := &cobra.Command{
		Use:  "get-database",
		RunE: register(GetDatabase),
	}
	database.AddCommand(getDatabase)

	deleteDatabase := &cobra.Command{
		Use:  "delete-database",
		RunE: register(DeleteDatabase),
	}
	database.AddCommand(deleteDatabase)

	listDatabases := &cobra.Command{
		Use:  "list-databases",
		RunE: register(ListDatabases),
	}
	database.AddCommand(listDatabases)

	createTable := &cobra.Command{
		Use:  "create-table <name>",
		RunE: register(CreateTable),
		Args: cobra.ExactArgs(1),
	}
	database.AddCommand(createTable)

	listTables := &cobra.Command{
		Use:  "list-tables",
		RunE: register(ListTables),
	}
	database.AddCommand(listTables)

	deleteTable := &cobra.Command{
		Use:  "delete-table <name>",
		RunE: register(DeleteTable),
		Args: cobra.ExactArgs(1),
	}
	database.AddCommand(deleteTable)

	rootCmd.AddCommand(database)
}

func getDbConfig(args []string) (*database.Config, error) {
	switch args[1] {
	case "postgres":
		return &database.Config{
			Config: &database.Config_Postgres{
				Postgres: &database.PostgresConfig{
					Credentials: &model.ProviderCredentials{
						PublicKey:  args[2],
						PrivateKey: args[3],
					},
					Host: args[4],
				},
			},
		}, nil
	case "mongo":
		return &database.Config{
			Config: &database.Config_Mongo{
				Mongo: &database.MongoConfig{
					Credentials: &model.ProviderCredentials{
						PublicKey:  args[2],
						PrivateKey: args[3],
					},
					Host: args[4],
				},
			},
		}, nil
	default:
		return nil, errors.New("invalid database type")
	}
}

func CreateDatabase(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	databaseName := args[0]
	if databaseName == "" {
		return errors.New("database name is required")
	}
	config, err := getDbConfig(args)
	if err != nil {
		return err
	}
	db, err := ds.Create(ctx, databaseName, config, false)
	if err != nil {
		return err
	}
	logger.Info("created database", zap.String("name", databaseName), zap.String("id", db.GetDatabaseId()))
	return nil
}

func GetDatabase(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	if databaseId == "" {
		return errors.New("missing required database id")
	}

	dbId, err := uuid.Parse(databaseId)
	if err != nil {
		return err
	}

	database, _, err := ds.Get(ctx, dbId)
	if err != nil {
		return err
	}
	logger.Info("got database", zap.String("name", database.Name))
	return nil
}

func DeleteDatabase(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	if databaseId == "" {
		return errors.New("missing required database id")
	}

	dbId, err := uuid.Parse(databaseId)
	if err != nil {
		return err
	}

	if err := ds.Delete(ctx, dbId); err != nil {
		return err
	}
	logger.Info("deleted database")
	return nil
}

func ListDatabases(ctx context.Context, cmd *cobra.Command, ds *service_database.Service, logger *zap.Logger) error {
	it, count, err := ds.List(ctx, &model.Pagination{})
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("found %d databases: ", count))
	for {
		database, err := it.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		logger.Info("", zap.String("name", database.Name), zap.Time("createdAt", database.GetCreatedAt().AsTime()))
	}
	return nil
}

func CreateTable(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	if databaseId == "" {
		return errors.New("missing required database id")
	}
	dbId, err := uuid.Parse(databaseId)
	if err != nil {
		return err
	}

	tableName := args[0]
	if tableName == "" {
		return errors.New("table name is required")
	}
	if err := ds.CreateTable(ctx, dbId, tableName); err != nil {
		return err
	}
	logger.Info("created table", zap.String("name", tableName))
	return nil
}

func ListTables(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	if databaseId == "" {
		return errors.New("missing required database id")
	}
	dbId, err := uuid.Parse(databaseId)
	if err != nil {
		return err
	}

	tables, err := ds.ListTables(ctx, dbId)
	if err != nil {
		return err
	}
	for _, table := range tables {
		logger.Info("table", zap.String("name", table.GetName()))
	}
	return nil
}

func DeleteTable(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	if databaseId == "" {
		return errors.New("missing required database id")
	}
	dbId, err := uuid.Parse(databaseId)
	if err != nil {
		return err
	}

	tableName := args[0]
	if tableName == "" {
		return errors.New("table name is required")
	}
	if err := ds.DeleteTable(ctx, dbId, tableName); err != nil {
		return err
	}
	logger.Info("deleted table", zap.String("name", tableName))
	return nil
}
