package database

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) delete(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	db, id, err := common.GetDatabase(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	if _, err := c.Rig.Database().Delete(ctx, &connect.Request[database.DeleteRequest]{
		Msg: &database.DeleteRequest{
			DatabaseId: id,
		},
	}); err != nil {
		return err
	}
	cmd.Println("deleted database", db.GetName())
	return nil
}
