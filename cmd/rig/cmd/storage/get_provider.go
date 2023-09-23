package storage

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) getProvider(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	g, uid, err := common.GetStorageProvider(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	if outputJson {
		cmd.Println(common.ProtoToPrettyJson(g))
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
