package database

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func ListTables(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := utils.GetDatabase(ctx, identifier, nc)
	if err != nil {
		return err
	}

	res, err := nc.Database().ListTables(ctx, &connect.Request[database.ListTablesRequest]{
		Msg: &database.ListTablesRequest{
			DatabaseId: id,
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
		for _, tb := range res.Msg.GetTables() {
			cmd.Println(utils.ProtoToPrettyJson(tb))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Tables (%d)", res.Msg.GetTotal()), "Name"})
	for i, tb := range res.Msg.GetTables() {
		t.AppendRow(table.Row{i + 1, tb.GetName()})
	}
	cmd.Println(t.Render())
	return nil
}
