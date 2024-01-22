package auth

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, _ []string) error {
	res, err := c.Rig.Authentication().Get(ctx, &connect.Request[authentication.GetRequest]{
		Msg: &authentication.GetRequest{},
	})
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(res.Msg, flags.Flags.OutputType)
	}

	ui := res.Msg.GetUserInfo()

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Entry", "Value"})
	t.AppendRow(table.Row{"ID", res.Msg.GetUserId()})
	t.AppendRow(table.Row{"Username", ui.GetUsername()})
	t.AppendRow(table.Row{"Email", ui.GetEmail()})
	t.AppendRow(table.Row{"Phone number", ui.GetPhoneNumber()})
	t.AppendRow(table.Row{"Created at", ui.GetCreatedAt().AsTime().Format(time.RFC3339)})

	cmd.Println(t.Render())

	return nil
}
