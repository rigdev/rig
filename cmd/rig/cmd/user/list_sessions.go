package user

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) listSessions(ctx context.Context, cmd *cobra.Command, args []string) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := common.GetUser(ctx, identifier, c.Rig, c.Prompter)
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

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetSessions(), flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{
		fmt.Sprintf("Sessions (%d)", resp.Msg.GetTotal()),
		"Session-Id",
		"Auth Method",
		"Device",
	})
	for i, s := range resp.Msg.GetSessions() {
		t.AppendRow(table.Row{
			i + 1,
			s.GetSessionId(),
			s.GetSession().GetAuthMethod().GetMethod(),
		})
	}

	cmd.Println(t.Render())

	return nil
}
