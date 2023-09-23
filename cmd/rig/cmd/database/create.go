package database

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) create(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	var err error
	if name == "" {
		name, err = common.PromptInput("Database name:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	if dbTypeString == "" {
		_, dbTypeString, err = common.PromptSelect("Database type:", []string{"mongo", "postgres"})
		if err != nil {
			return err
		}
	}

	dbType, err := getDbType(dbTypeString)
	if err != nil {
		return err
	}

	res, err := c.Rig.Database().Create(ctx, &connect.Request[database.CreateRequest]{Msg: &database.CreateRequest{
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
