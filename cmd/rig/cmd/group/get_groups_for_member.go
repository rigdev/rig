package group

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
)

func (c *Cmd) listGroupsForUser(ctx context.Context, cmd *cobra.Command, args []string) error {
	var memberID *group.MemberID
	var err error
	if len(args) == 0 {
		userID, serviceAccountID, _, err := common.GetMember(ctx, c.Rig)
		if err != nil {
			return err
		}
		if userID != "" {
			memberID = &group.MemberID{
				MemberID: &group.MemberID_UserId{
					UserId: userID,
				},
			}
		} else if serviceAccountID != "" {
			memberID = &group.MemberID{
				MemberID: &group.MemberID_ServiceAccountId{
					ServiceAccountId: serviceAccountID,
				},
			}
		}
	}

	resp, err := c.Rig.Group().ListGroupsForMember(ctx, &connect.Request[group.ListGroupsForMemberRequest]{
		Msg: &group.ListGroupsForMemberRequest{
			MemberId: memberID,
			Pagination: &model.Pagination{
				Offset: uint32(offset),
				Limit:  uint32(limit),
			},
		},
	})
	if err != nil {
		return err
	}

	if base.Flags.OutputType != base.OutputTypePretty {
		return base.FormatPrint(resp.Msg.GetGroups())
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Groups (%d)", resp.Msg.GetTotal()), "ID"})
	for i, g := range resp.Msg.GetGroups() {
		t.AppendRow(table.Row{i + 1, g.GetGroupId()})
	}
	cmd.Println(t.Render())

	return nil
}
