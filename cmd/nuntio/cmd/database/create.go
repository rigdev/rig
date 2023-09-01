package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func Create(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	var err error
	if name == "" {
		name, err = utils.PromptGetInput("Database name:", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}

	if dbTypeString == "" {
		_, dbTypeString, err = utils.PromptSelect("Database type:", []string{"mongo", "postgres"}, false)
		if err != nil {
			return err
		}
	}

	dbType, err := getDbType(dbTypeString)
	if err != nil {
		return err
	}

	res, err := nc.Database().Create(ctx, &connect.Request[database.CreateRequest]{Msg: &database.CreateRequest{
		Initializers: []*database.Update{
			{Field: &database.Update_Name{Name: name}},
		},
		Type: dbType,
	}})
	if err != nil {
		return err
	}

	cmd.Printf("created database %s of type %s with id %s\n", name, dbTypeString, res.Msg.GetDatabase().GetDatabaseId())
	return nil
}
