package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func UserAddMember(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	uidentifier := ""
	if len(args) > 1 {
		uidentifier = args[0]
	}
	u, uuid, err := common.GetUser(ctx, uidentifier, nc)
	if err != nil {
		return err
	}

	var guid string
	var gname string
	if groupIdentifier == "" {
		groupsRes, err := nc.Group().List(ctx, connect.NewRequest(&group.ListRequest{}))
		if err != nil {
			return err
		}

		gs := groupsRes.Msg.GetGroups()
		ugs := u.GetUserInfo().GetGroupIds()
		gsMap := make(map[string]string)
		for _, g := range gs {
			gsMap[g.GetGroupId()] = g.GetName()
		}

		for _, g := range ugs {
			delete(gsMap, g)
		}

		if len(gsMap) == 0 {
			cmd.Println("No groups available")
			return nil
		}

		var gnames []string
		for _, gname := range gsMap {
			gnames = append(gnames, gname)
		}

		_, gname, err := common.PromptSelect("Select group", gnames, false)
		if err != nil {
			return err
		}

		var ok bool
		guid, ok = gsMap[gname]
		if !ok {
			return nil
		}
	} else {
		g, id, err := common.GetGroup(ctx, groupIdentifier, nc)
		if err != nil {
			return err
		}
		guid = id
		gname = g.GetName()
	}

	_, err = nc.Group().AddMember(ctx, &connect.Request[group.AddMemberRequest]{
		Msg: &group.AddMemberRequest{
			GroupId: guid,
			UserIds: []string{uuid},
		},
	})
	if err != nil {
		return err
	}

	cmd.Printf("User added to group %s\n", gname)
	return nil
}
