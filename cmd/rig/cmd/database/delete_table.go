package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func DeleteTable(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := utils.GetDatabase(ctx, identifier, nc)
	if err != nil {
		return err
	}

	if name == "" {
		name, err = utils.PromptGetInput("Table Name", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}

	if _, err := nc.Database().DeleteTable(ctx, &connect.Request[database.DeleteTableRequest]{
		Msg: &database.DeleteTableRequest{
			DatabaseId: id,
			TableName:  name,
		},
	}); err != nil {
		return err
	}
	cmd.Println("deleted table\n", name)
	return nil
}
