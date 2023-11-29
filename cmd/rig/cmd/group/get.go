package group

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		identifier := ""
		if len(args) > 0 {
			identifier = args[0]
		}
		g, uid, err := common.GetGroup(ctx, identifier, c.Rig)
		if err != nil {
			return err
		}

		if base.Flags.OutputType != base.OutputTypePretty {
			return base.FormatPrint(g)
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

	if base.Flags.OutputType != base.OutputTypePretty {
		return base.FormatPrint(resp.Msg.GetGroups())
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Groups (%d)", resp.Msg.GetTotal()), "ID"})
	for i, g := range resp.Msg.GetGroups() {
		t.AppendRow(table.Row{i + 1, g.GetGroupId()})
	}
	cmd.Println(t.Render())
	return nil
}
