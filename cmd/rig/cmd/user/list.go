package user

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/user"
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
		u, id, err := common.GetUser(ctx, identifier, c.Rig)
		if err != nil {
			return err
		}

		if base.Flags.OutputType != base.OutputTypePretty {
			return base.FormatPrint(u)
		}

		t := table.NewWriter()
		t.AppendHeader(table.Row{"Attribute", "Value"})
		t.AppendRows([]table.Row{
			{"ID", id},
			{"Name", u.GetProfile().GetFirstName() + " " + u.GetProfile().GetLastName()},
			{"Email", u.GetUserInfo().GetEmail()},
			{"Phone number", u.GetUserInfo().GetPhoneNumber()},
			{"Username", u.GetUserInfo().GetUsername()},
			{"First name", u.GetProfile().GetFirstName()},
			{"Last name", u.GetProfile().GetLastName()},
			{"Is Email verified", u.GetIsEmailVerified()},
			{"Is Phone verified", u.GetIsPhoneVerified()},
			{"Created at", u.GetUserInfo().GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")},
		})
		cmd.Println(t.Render())
		return nil
	}

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

	if base.Flags.OutputType != base.OutputTypePretty {
		return base.FormatPrint(resp.Msg.GetUsers())
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Users (%d)", resp.Msg.GetTotal()), "Identifier", "ID"})
	for i, u := range resp.Msg.GetUsers() {
		t.AppendRow(table.Row{i + 1, u.GetPrintableName(), u.GetUserId()})
	}
	cmd.Println(t.Render())
	return nil
}
