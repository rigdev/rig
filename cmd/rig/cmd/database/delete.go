package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func Delete(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	db, id, err := common.GetDatabase(ctx, identifier, nc)
	if err != nil {
		return err
	}

	if _, err := nc.Database().Delete(ctx, &connect.Request[database.DeleteRequest]{
		Msg: &database.DeleteRequest{
			DatabaseId: id,
		},
	}); err != nil {
		return err
	}
	cmd.Println("deleted database", db.GetName())
	return nil
}
