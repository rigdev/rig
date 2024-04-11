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

func (c *Cmd) list(ctx context.Context, cmd *cobra.Command, _ []string) error {
	req := &user.ListRequest{
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
	}
	resp, err := c.Rig.User().List(ctx, &connect.Request[user.ListRequest]{Msg: req})
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetUsers(), flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Users (%d)", resp.Msg.GetTotal()), "Identifier", "ID"})
	for i, u := range resp.Msg.GetUsers() {
		t.AppendRow(table.Row{i + 1, u.GetPrintableName(), u.GetUserId()})
	}
	cmd.Println(t.Render())
	return nil
}
