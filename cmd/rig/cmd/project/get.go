package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, args []string) error {
	req := &project.GetRequest{}
	resp, err := c.Rig.Project().Get(ctx, &connect.Request[project.GetRequest]{Msg: req})
	if err != nil {
		return err
	}

	if outputJSON {
		cmd.Println(common.ProtoToPrettyJson(resp.Msg.GetProject()))
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRows([]table.Row{
		{"Name", resp.Msg.GetProject().GetName()},
		{"Installation ID", resp.Msg.Project.GetInstallationId()},
		{"Created At", resp.Msg.GetProject().GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")},
	})

	cmd.Println(t.Render())
	return nil
}
