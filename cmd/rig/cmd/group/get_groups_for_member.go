package group

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/pkg/errors"
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
				Kind: &group.MemberID_UserId{
					UserId: userID,
				},
			}
		} else if serviceAccountID != "" {
			memberID = &group.MemberID{
				Kind: &group.MemberID_ServiceAccountId{
					ServiceAccountId: serviceAccountID,
				},
			}
		}
	} else {
		_, err := c.Rig.User().Get(ctx, &connect.Request[user.GetRequest]{
			Msg: &user.GetRequest{
				UserId: args[0],
			},
		})
		if err == nil {
			memberID = &group.MemberID{
				Kind: &group.MemberID_UserId{
					UserId: args[0],
				},
			}
		} else {
			accs, err := c.Rig.ServiceAccount().List(ctx, &connect.Request[service_account.ListRequest]{
				Msg: &service_account.ListRequest{},
			})
			if err != nil {
				return err
			}
			for _, acc := range accs.Msg.GetServiceAccounts() {
				if acc.GetServiceAccountId() == args[0] {
					memberID = &group.MemberID{
						Kind: &group.MemberID_ServiceAccountId{
							ServiceAccountId: args[0],
						},
					}
					break
				}
			}
			if memberID == nil {
				return errors.InvalidArgumentErrorf("unknown member %q", args[0])
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
