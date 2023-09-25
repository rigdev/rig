package database

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) get(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	db, id, err := common.GetDatabase(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	if outputJSON {
		cmd.Println(common.ProtoToPrettyJson(db))
		return nil
	}

	// print a table with a column for attributes, and a column for values
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRows([]table.Row{
		{"ID", id},
		{"Name", db.GetName()},
		{"Type", db.GetType()},
		{"Num Tables", len(db.GetInfo().GetTables())},
		{"Num Creds", len(db.GetInfo().GetCredentials())},
		{"Created At", db.GetInfo().GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")},
	})
	cmd.Println(t.Render())

	return nil
}
