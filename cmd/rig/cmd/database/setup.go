package database

import (
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
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
	clientID     string
	clientSecret string
)

func Setup(parent *cobra.Command) {
	database := &cobra.Command{
		Use: "database",
	}

	getDatabaseEndpoint := &cobra.Command{
		Use:  "get-endpoint [id | name]",
		RunE: base.Register(GetEndpoint),
		Args: cobra.MaximumNArgs(1),
	}
	getDatabaseEndpoint.Flags().StringVarP(&clientID, "client-id", "i", "", "client id")
	getDatabaseEndpoint.Flags().StringVarP(&clientSecret, "client-secret", "s", "", "client secret")
	database.AddCommand(getDatabaseEndpoint)

	createDatabase := &cobra.Command{
		Use:  "create",
		RunE: base.Register(Create),
		Args: cobra.NoArgs,
	}
	createDatabase.Flags().StringVarP(&name, "name", "n", "", "name of the database")
	database.AddCommand(createDatabase)

	createDatabaseCredentials := &cobra.Command{
		Use:  "create-credentials [id | db-name]",
		RunE: base.Register(CreateCredentials),
		Args: cobra.NoArgs,
	}
	database.AddCommand(createDatabaseCredentials)

	deleteDatabaseCredentials := &cobra.Command{
		Use:  "delete-credentials [id | db-name]",
		RunE: base.Register(DeleteCredentials),
		Args: cobra.MaximumNArgs(1),
	}
	deleteDatabaseCredentials.Flags().StringVarP(&clientID, "clientId", "n", "", "clientId of the credentials")
	database.AddCommand(deleteDatabaseCredentials)

	getDatabase := &cobra.Command{
		Use:  "get [id | name]",
		RunE: base.Register(Get),
		Args: cobra.MaximumNArgs(1),
	}
	getDatabase.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	database.AddCommand(getDatabase)

	delete := &cobra.Command{
		Use:  "delete [id | name]",
		RunE: base.Register(Delete),
		Args: cobra.MaximumNArgs(1),
	}
	database.AddCommand(delete)

	list := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		RunE:    base.Register(List),
		Args:    cobra.NoArgs,
	}
	list.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	list.Flags().IntVarP(&offset, "offset", "o", 0, "offset")
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	database.AddCommand(list)

	createTable := &cobra.Command{
		Use:  "create-table [id | db-name]",
		RunE: base.Register(CreateTable),
		Args: cobra.MaximumNArgs(1),
	}
	createTable.Flags().StringVarP(&name, "name", "n", "", "name of the table")
	database.AddCommand(createTable)

	listTables := &cobra.Command{
		Use:  "list-tables [id | name]",
		RunE: base.Register(ListTables),
		Args: cobra.MaximumNArgs(1),
	}
	listTables.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	listTables.Flags().IntVarP(&offset, "offset", "o", 0, "offset")
	listTables.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	database.AddCommand(listTables)

	deleteTable := &cobra.Command{
		Use:  "delete-table [id | db-name]",
		RunE: base.Register(DeleteTable),
		Args: cobra.MaximumNArgs(1),
	}
	deleteTable.Flags().StringVarP(&name, "name", "n", "", "name of the table")
	database.AddCommand(deleteTable)

	parent.AddCommand(database)
}
