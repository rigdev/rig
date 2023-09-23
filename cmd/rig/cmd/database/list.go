package database

import (
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) list(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	res, err := c.Rig.Database().List(ctx, &connect.Request[database.ListRequest]{
		Msg: &database.ListRequest{
			Pagination: &model.Pagination{
				Offset: uint32(offset),
				Limit:  uint32(limit),
			},
		},
	})
	if err != nil {
		return err
	}

	if outputJSON {
		for _, db := range res.Msg.GetDatabases() {
			cmd.Println(common.ProtoToPrettyJson(db))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("DBs (%d)", res.Msg.GetTotal()), "Name", "Type"})
	for i, db := range res.Msg.GetDatabases() {
		t.AppendRow(table.Row{i + 1, db.GetName(), db.GetType()})
	}
	cmd.Println(t.Render())
	return nil
}
