package group

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func GroupListGroupsForUser(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, uid, err := common.GetUser(ctx, identifier, nc)
	if err != nil {
		return err
	}

	resp, err := nc.Group().ListGroupsForUser(ctx, &connect.Request[group.ListGroupsForUserRequest]{
		Msg: &group.ListGroupsForUserRequest{
			UserId: uid,
			Pagination: &model.Pagination{
				Offset: uint32(offset),
				Limit:  uint32(limit),
			},
		},
	})
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
