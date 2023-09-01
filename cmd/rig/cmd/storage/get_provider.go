package storage

import (
	"context"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func StorageGetProvider(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	g, uid, err := utils.GetStorageProvider(ctx, identifier, nc)
	if err != nil {
		return err
	}

	if outputJson {
		cmd.Println(utils.ProtoToPrettyJson(g))
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRows([]table.Row{
		{"ID", uid},
		{"Name", g.GetName()},
		{"#Buckets", len(g.GetBuckets())},
		{"Created at", g.GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")},
	})
	cmd.Println(t.Render())
	return nil
}
