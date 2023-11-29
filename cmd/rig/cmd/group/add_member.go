package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) addMember(ctx context.Context, cmd *cobra.Command, args []string) error {
	var memberID string
	var groupIDs []string
	var err error
	if len(args) == 0 {
		memberID, groupIDs, err = common.GetMember(ctx, c.Rig)
		if err != nil {
			return err
		}
	}

	var gname string
	if groupID == "" {
		groupsRes, err := c.Rig.Group().List(ctx, connect.NewRequest(&group.ListRequest{}))
		if err != nil {
			return err
		}

		gs := groupsRes.Msg.GetGroups()
		gsMap := make(map[string]string)
		for _, g := range gs {
			gsMap[g.GetGroupId()] = ""
		}

		for _, g := range groupIDs {
			delete(gsMap, g)
		}

		if len(gsMap) == 0 {
			cmd.Println("No groups available")
			return nil
		}

		var gIDs []string
		for gID := range gsMap {
			gIDs = append(gIDs, gID)
		}

		_, groupID, err = common.PromptSelect("Select group", gIDs)
		if err != nil {
			return err
		}
	} else {
		_, groupID, err = common.GetGroup(ctx, groupID, c.Rig)
		if err != nil {
			return err
		}
	}

	_, err = c.Rig.Group().AddMember(ctx, &connect.Request[group.AddMemberRequest]{
		Msg: &group.AddMemberRequest{
			GroupId:   groupID,
			MemberIds: []string{memberID},
		},
	})
	if err != nil {
		return err
	}

	cmd.Printf("User added to group %s\n", gname)
	return nil
}
