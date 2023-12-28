package project

import (
	"context"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, _ []string) error {
	req := &project.GetRequest{
		ProjectId: c.Cfg.GetProject(),
	}
	resp, err := c.Rig.Project().Get(ctx, &connect.Request[project.GetRequest]{
		Msg: req,
	})
	if err != nil {
		return err
	}

	if base.Flags.OutputType != base.OutputTypePretty {
		return base.FormatPrint(resp.Msg.GetProject())
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
