package project

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
)

func (c *Cmd) list(ctx context.Context, cmd *cobra.Command, _ []string) error {
	req := &project.ListRequest{
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
	}

	resp, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{Msg: req})
	if err != nil {
		return err
	}

	if base.Flags.OutputType != base.OutputTypePretty {
		return base.FormatPrint(resp.Msg.GetProjects())
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Projects (%d)", resp.Msg.GetTotal()), "Name", "ID", "Created At"})
	for i, p := range resp.Msg.GetProjects() {
		t.AppendRow(table.Row{i + 1, p.GetName(), p.GetProjectId(), p.GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")})
	}
	cmd.Println(t.Render())
	return nil
}
