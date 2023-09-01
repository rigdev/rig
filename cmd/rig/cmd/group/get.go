package group

import (
	"context"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func GroupGet(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	g, uid, err := utils.GetGroup(ctx, identifier, nc)
	if err != nil {
		return err
	}

	if outputJSON {
		cmd.Println(utils.ProtoToPrettyJson(g))
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRows([]table.Row{
		{"ID", uid},
		{"Name", g.GetName()},
		{"#Members", g.GetNumMembers()},
		{"Created at", g.GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")},
	})
	cmd.Println(t.Render())
	return nil
}
