package group

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) list(ctx context.Context, cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		identifier := ""
		if len(args) > 0 {
			identifier = args[0]
		}
		g, uid, err := common.GetGroup(ctx, identifier, c.Rig, c.Prompter)
		if err != nil {
			return err
		}

		if flags.Flags.OutputType != common.OutputTypePretty {
			return common.FormatPrint(g, flags.Flags.OutputType)
		}

		t := table.NewWriter()
		t.AppendHeader(table.Row{"Attribute", "Value"})
		t.AppendRows([]table.Row{
			{"ID", uid},
			{"#Members", g.GetNumMembers()},
			{"Created at", g.GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")},
		})
		cmd.Println(t.Render())
		return nil
	}

	req := &group.ListRequest{
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
	}
	resp, err := c.Rig.Group().List(ctx, &connect.Request[group.ListRequest]{Msg: req})
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetGroups(), flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Groups (%d)", resp.Msg.GetTotal()), "ID"})
	for i, g := range resp.Msg.GetGroups() {
		t.AppendRow(table.Row{i + 1, g.GetGroupId()})
	}
	cmd.Println(t.Render())
	return nil
}
