package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func UserRemoveMember(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	uidentifier := ""
	if len(args) > 1 {
		uidentifier = args[1]
	}
	_, uuid, err := common.GetUser(ctx, uidentifier, nc)
	if err != nil {
		return err
	}
	var guid string
	var gname string
	if groupIdentifier == "" {
		res, err := nc.Group().ListGroupsForUser(ctx, &connect.Request[group.ListGroupsForUserRequest]{
			Msg: &group.ListGroupsForUserRequest{
				UserId: uuid,
			},
		})
		if err != nil {
			return err
		}
		fields := make([]string, len(res.Msg.GetGroups()))
		for i, g := range res.Msg.GetGroups() {
			fields[i] = g.GetName()
		}

		var i int
		i, gname, err = common.PromptSelect("Select group", fields)
		if err != nil {
			return err
		}

		guid = res.Msg.GetGroups()[i].GetGroupId()
	} else {
		g, id, err := common.GetGroup(ctx, groupIdentifier, nc)
		if err != nil {
			return err
		}
		guid = id
		gname = g.GetName()
	}

	_, err = nc.Group().RemoveMember(ctx, &connect.Request[group.RemoveMemberRequest]{
		Msg: &group.RemoveMemberRequest{
			GroupId: guid,
			UserId:  uuid,
		},
	})
	if err != nil {
		return err
	}

	cmd.Printf("User removed from group %s\n", gname)
	return nil
}
