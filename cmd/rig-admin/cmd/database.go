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
		Use:  "create-database <db> <type>",
		RunE: register(CreateDatabase),
		Args: cobra.ExactArgs(2),
	}
	database.AddCommand(createDatabase)

	createDatabaseCredential := &cobra.Command{
		Use:  "create-credential <name>",
		RunE: register(CreateDatabaseCredential),
		Args: cobra.ExactArgs(1),
	}
	database.AddCommand(createDatabaseCredential)

	deleteDatabaseCredential := &cobra.Command{
		Use:  "delete-credential <clientId>",
		RunE: register(DeleteDatabaseCredential),
		Args: cobra.ExactArgs(1),
	}
	database.AddCommand(deleteDatabaseCredential)

	listCredentials := &cobra.Command{
		Use:  "list-credentials",
		RunE: register(ListDatabaseCredentials),
	}
	database.AddCommand(listCredentials)

	getDatabase := &cobra.Command{
		Use:  "get-database",
		RunE: register(GetDatabase),
	}
	database.AddCommand(getDatabase)

	connectDatabase := &cobra.Command{
		Use:  "connect <clientId> <clientSecret>",
		RunE: register(ConnectDatabase),
		Args: cobra.ExactArgs(2),
	}
	database.AddCommand(connectDatabase)

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

func getDbType(databaseType string) (database.Type, error) {
	if databaseType == "" {
		return database.Type_TYPE_UNSPECIFIED, errors.New("type is required")
	}
	var dbType database.Type
	switch databaseType {
	case "mongo":
		dbType = database.Type_TYPE_MONGO
	case "postgres":
		dbType = database.Type_TYPE_POSTGRES
	default:
		return database.Type_TYPE_UNSPECIFIED, fmt.Errorf("invalid database type: %v (insert mongo or postgres)", databaseType)
	}
	return dbType, nil
}

func CreateDatabase(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	databaseName := args[0]
	if databaseName == "" {
		return errors.New("database name is required")
	}
	dbType, err := getDbType(args[1])
	if err != nil {
		return err
	}
	databaseID, db, err := ds.Create(ctx, dbType, []*database.Update{
		{Field: &database.Update_Name{Name: databaseName}},
	})
	if err != nil {
		return err
	}
	logger.Info("created database", zap.String("name", databaseName), zap.String("type", db.GetType().String()), zap.String("id", databaseID.String()))
	return nil
}

func CreateDatabaseCredential(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	if databaseId == "" {
		return errors.New("missing required database id")
	}
	credentialName := args[0]
	if credentialName == "" {
		return errors.New("credential name is required")
	}

	dbId, err := uuid.Parse(databaseId)
	if err != nil {
		return err
	}

	clientId, clientSecret, err := ds.CreateCredential(ctx, credentialName, dbId)
	if err != nil {
		return err
	}
	logger.Info("created credential", zap.String("clientId", clientId), zap.String("secret", clientSecret))
	return nil
}

func DeleteDatabaseCredential(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	if databaseId == "" {
		return errors.New("missing required database id")
	}
	credentialsName := args[0]
	if credentialsName == "" {
		return errors.New("credentials name is required")
	}

	dbId, err := uuid.Parse(databaseId)
	if err != nil {
		return err
	}

	if err := ds.DeleteCredential(ctx, credentialsName, dbId); err != nil {
		return err
	}
	logger.Info("deleted credential", zap.String("clientId", credentialsName))
	return nil
}

func ListDatabaseCredentials(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	if databaseId == "" {
		return errors.New("missing required database id")
	}

	dbId, err := uuid.Parse(databaseId)
	if err != nil {
		return err
	}

	credentials, err := ds.ListCredentials(ctx, dbId)
	if err != nil {
		return err
	}
	logger.Info("found credentials", zap.Int("amount", len(credentials)))
	for _, credential := range credentials {
		logger.Info("credential", zap.String("name", credential.GetName()), zap.String("clientId", credential.ClientId), zap.String("secret", string(credential.Secret)))
	}
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

	database, err := ds.Get(ctx, dbId)
	if err != nil {
		return err
	}
	logger.Info("got database", zap.String("name", database.Name))
	return nil
}

func ConnectDatabase(ctx context.Context, cmd *cobra.Command, args []string, ds *service_database.Service, logger *zap.Logger) error {
	if databaseId == "" {
		return errors.New("missing required database id")
	}

	dbId, err := uuid.Parse(databaseId)
	if err != nil {
		return err
	}

	clientId := args[0]
	if clientId == "" {
		return errors.New("clientId is required")
	}
	clientSecret := args[1]
	if clientSecret == "" {
		return errors.New("clientSecret is required")
	}
	endpoint, _, err := ds.GetDatabaseEndpoint(ctx, dbId, clientId, clientSecret)
	if err != nil {
		return err
	}
	fmt.Printf(`

connect to mongo by running:
mongosh "%s"

	`, endpoint)

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
		logger.Info("", zap.String("name", database.Name), zap.String("type", database.GetType().String()), zap.Time("createdAt", database.GetInfo().GetCreatedAt().AsTime()))
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
