package group

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
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

		if outputJSON {
			cmd.Println(common.ProtoToPrettyJson(g))
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

	if outputJSON {
		for _, u := range resp.Msg.GetGroups() {
			cmd.Println(common.ProtoToPrettyJson(u))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Groups (%d)", resp.Msg.GetTotal()), "Name", "ID"})
	for i, g := range resp.Msg.GetGroups() {
		t.AppendRow(table.Row{i + 1, g.GetName(), g.GetGroupId()})
	}
	cmd.Println(t.Render())
	return nil
}
