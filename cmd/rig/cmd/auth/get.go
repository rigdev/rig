package auth

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, args []string) error {
	res, err := c.Rig.Authentication().Get(ctx, &connect.Request[authentication.GetRequest]{
		Msg: &authentication.GetRequest{},
	})
	if err != nil {
		return err
	}

	if base.Flags.OutputType != base.OutputTypePretty {
		return base.FormatPrint(res.Msg)
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
