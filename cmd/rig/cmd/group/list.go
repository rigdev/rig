package group

import (
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) list(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	search := strings.Join(args, " ")
	req := &group.ListRequest{
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
		Search: search,
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
