package project

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
)

func ProjectList(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client, cfg *cmd_config.Config) error {
	req := &project.ListRequest{
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
	}

	resp, err := nc.Project().List(ctx, &connect.Request[project.ListRequest]{Msg: req})
	if err != nil {
		return err
	}

	if outputJSON {
		for _, p := range resp.Msg.GetProjects() {
			cmd.Println(common.ProtoToPrettyJson(p))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Projects (%d)", resp.Msg.GetTotal()), "Name", "ID", "Created At"})
	for i, p := range resp.Msg.GetProjects() {
		t.AppendRow(table.Row{i + 1, p.GetName(), p.GetProjectId(), p.GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")})
	}
	cmd.Println(t.Render())
	return nil
}
