package user

import (
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) list(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	search := strings.Join(args, " ")
	req := &user.ListRequest{
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
		Search: search,
	}
	resp, err := c.Rig.User().List(ctx, &connect.Request[user.ListRequest]{Msg: req})
	if err != nil {
		return err
	}

	if outputJson {
		for _, u := range resp.Msg.GetUsers() {
			cmd.Println(common.ProtoToPrettyJson(u))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Users (%d)", resp.Msg.GetTotal()), "Identifier", "ID"})
	for i, u := range resp.Msg.GetUsers() {
		t.AppendRow(table.Row{i + 1, u.GetPrintableName(), u.GetUserId()})
	}
	cmd.Println(t.Render())
	return nil
}
