package database

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) createTable(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := common.GetDatabase(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	if name == "" {
		name, err = common.PromptInput("Table Name:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	if _, err := c.Rig.Database().CreateTable(ctx, &connect.Request[database.CreateTableRequest]{
		Msg: &database.CreateTableRequest{
			DatabaseId: id,
			TableName:  name,
		},
	}); err != nil {
		return err
	}

	cmd.Printf("created table %s\n", name)
	return nil
}
