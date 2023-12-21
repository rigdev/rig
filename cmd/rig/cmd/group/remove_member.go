package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) removeMember(ctx context.Context, cmd *cobra.Command, args []string) error {
	var userID string
	var serviceAccountID string
	var err error
	if len(args) == 0 {
		userID, serviceAccountID, _, err = common.GetMember(ctx, c.Rig)
		if err != nil {
			return err
		}
	}

	var memberID *group.MemberID
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
