package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	outputJSON bool
)

var (
	offset int
	limit  int
)

var (
	name         string
	dbTypeString string
	clientID     string
	clientSecret string
)

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
	Cfg *cmd_config.Config
}

func (c Cmd) Setup(parent *cobra.Command) {
	database := &cobra.Command{
		Use: "database",
	}

	createDatabase := &cobra.Command{
		Use:  "create",
		RunE: c.create,
		Args: cobra.NoArgs,
	}
	createDatabase.Flags().StringVarP(&name, "name", "n", "", "name of the database")
	createDatabase.Flags().StringVarP(&dbTypeString, "type", "t", "", "type of the database (mongo, postgres)")
	database.AddCommand(createDatabase)

	createDatabaseCredential := &cobra.Command{
		Use:  "create-credentials [id | db-name]",
		RunE: c.createCredentials,
		Args: cobra.MaximumNArgs(1),
	}
	createDatabaseCredential.Flags().StringVarP(&name, "name", "n", "", "name of the credentials")
	database.AddCommand(createDatabaseCredential)

	deleteCredential := &cobra.Command{
		Use:  "delete-credentials [id | db-name]",
		RunE: c.deleteCredential,
		Args: cobra.MaximumNArgs(1),
	}
	deleteCredential.Flags().StringVarP(&name, "name", "n", "", "name of the credentials")
	database.AddCommand(deleteCredential)

	getDatabase := &cobra.Command{
		Use:  "get [id | name]",
		RunE: c.get,
		Args: cobra.MaximumNArgs(1),
	}
	getDatabase.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	database.AddCommand(getDatabase)

	connect := &cobra.Command{
		Use:  "connect [id | name]",
		RunE: c.connect,
		Args: cobra.MaximumNArgs(1),
	}
	connect.Flags().StringVarP(&clientID, "client-id", "i", "", "client id")
	connect.Flags().StringVarP(&clientSecret, "client-secret", "s", "", "client secret")
	database.AddCommand(connect)

	delete := &cobra.Command{
		Use:  "delete [id | name]",
		RunE: c.delete,
		Args: cobra.MaximumNArgs(1),
	}
	database.AddCommand(delete)

	list := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		RunE:    c.list,
		Args:    cobra.NoArgs,
	}
	list.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	list.Flags().IntVarP(&offset, "offset", "o", 0, "offset")
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	database.AddCommand(list)

	createTable := &cobra.Command{
		Use:  "create-table [id | db-name]",
		RunE: c.createTable,
		Args: cobra.MaximumNArgs(1),
	}
	createTable.Flags().StringVarP(&name, "name", "n", "", "name of the table")
	database.AddCommand(createTable)

	listTables := &cobra.Command{
		Use:  "list-tables [id | name]",
		RunE: c.listTables,
		Args: cobra.MaximumNArgs(1),
	}
	listTables.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	listTables.Flags().IntVarP(&offset, "offset", "o", 0, "offset")
	listTables.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	database.AddCommand(listTables)

	deleteTable := &cobra.Command{
		Use:  "delete-table [id | db-name]",
		RunE: c.deleteTable,
		Args: cobra.MaximumNArgs(1),
	}
	deleteTable.Flags().StringVarP(&name, "name", "n", "", "name of the table")
	database.AddCommand(deleteTable)

	parent.AddCommand(database)
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
