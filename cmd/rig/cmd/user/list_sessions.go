package user

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) listSessions(ctx context.Context, cmd *cobra.Command, args []string) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := common.GetUser(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	resp, err := c.Rig.User().ListSessions(ctx, connect.NewRequest(&user.ListSessionsRequest{
		UserId: id,
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
	}))
	if err != nil {
		return err
	}

	if outputJson {
		for _, s := range resp.Msg.GetSessions() {
			cmd.Println(common.ProtoToPrettyJson(s))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Sessions (%d)", resp.Msg.GetTotal()), "Session-Id", "Auth Method", "Device"})
	for i, s := range resp.Msg.GetSessions() {
		t.AppendRow(table.Row{i + 1, s.GetSessionId(), s.GetSession().GetAuthMethod().GetMethod(), s.GetSession().GetDevice()})
	}

	cmd.Println(t.Render())

	return nil
}
