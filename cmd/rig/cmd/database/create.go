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

	res, err := nc.Database().Create(ctx, &connect.Request[database.CreateRequest]{Msg: &database.CreateRequest{
		Name:   name,
		DbType: database.Type_TYPE_MONGODB,
	}})
	if err != nil {
		return err
	}

	cmd.Printf("created database %s with default secret %s and id %s\n", name, res.Msg.GetClientSecret(), res.Msg.GetDatabase().GetDatabaseId())
	return nil
}
