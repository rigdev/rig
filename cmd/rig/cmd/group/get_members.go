package group

import (
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) listMembers(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, uid, err := common.GetGroup(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	resp, err := c.Rig.Group().ListMembers(ctx, &connect.Request[group.ListMembersRequest]{
		Msg: &group.ListMembersRequest{
			GroupId: uid,
		},
	})
	if err != nil {
		return err
	}

	if outputJSON {
		for _, u := range resp.Msg.GetMembers() {
			cmd.Println(common.ProtoToPrettyJson(u))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Members (%d)", resp.Msg.GetTotal()), "Identifier", "ID"})
	for i, m := range resp.Msg.GetMembers() {
		t.AppendRow(table.Row{i + 1, m.GetUser().GetPrintableName(), m.GetUser().GetUserId()})
	}
	cmd.Println(t.Render())
	return nil
}
