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

func (c *Cmd) listMembers(ctx context.Context, cmd *cobra.Command, args []string) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, uid, err := common.GetGroup(ctx, identifier, c.Rig, c.Prompter)
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

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetMembers(), flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Members (%d)", resp.Msg.GetTotal()), "Identifier", "ID", "Kind"})
	for i, m := range resp.Msg.GetMembers() {
		switch v := m.GetEntry().(type) {
		case *model.MemberEntry_User:
			t.AppendRow(table.Row{i + 1, v.User.GetPrintableName(), v.User.GetUserId(), "User"})
		case *model.MemberEntry_ServiceAccount:
			t.AppendRow(table.Row{i + 1, v.ServiceAccount.GetName(), v.ServiceAccount.GetServiceAccountId(), "ServiceAccount"})
		}
	}
	cmd.Println(t.Render())
	return nil
}
