package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) removeMember(ctx context.Context, cmd *cobra.Command, args []string) error {
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

	if groupID == "" {
		res, err := c.Rig.Group().ListGroupsForMember(ctx, &connect.Request[group.ListGroupsForMemberRequest]{
			Msg: &group.ListGroupsForMemberRequest{
				MemberId: memberID,
			},
		})
		if err != nil {
			return err
		}
		fields := make([]string, len(res.Msg.GetGroups()))
		for i, g := range res.Msg.GetGroups() {
			fields[i] = g.GetGroupId()
		}

		_, groupID, err = common.PromptSelect("Select group", fields)
		if err != nil {
			return err
		}
	} else {
		_, id, err := common.GetGroup(ctx, groupID, c.Rig)
		if err != nil {
			return err
		}
		groupID = id
	}

	req := &group.RemoveMemberRequest{
		GroupId:  groupID,
		MemberId: memberID,
	}

	_, err = c.Rig.Group().RemoveMember(ctx, &connect.Request[group.RemoveMemberRequest]{
		Msg: req,
	})
	if err != nil {
		return err
	}

	cmd.Printf("User removed from group %s\n", groupID)
	return nil
}
