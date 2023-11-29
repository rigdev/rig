package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) removeMember(ctx context.Context, cmd *cobra.Command, args []string) error {
	var memberID string
	var err error
	if len(args) == 0 {
		memberID, _, err = common.GetMember(ctx, c.Rig)
		if err != nil {
			return err
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
	_, err = c.Rig.Group().RemoveMember(ctx, &connect.Request[group.RemoveMemberRequest]{
		Msg: &group.RemoveMemberRequest{
			GroupId:  groupID,
			MemberId: memberID,
		},
	})
	if err != nil {
		return err
	}

	cmd.Printf("User removed from group %s\n", groupID)
	return nil
}
