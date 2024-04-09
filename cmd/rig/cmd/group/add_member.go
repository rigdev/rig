package group

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) addMember(ctx context.Context, cmd *cobra.Command, args []string) error {
	var memberID *group.MemberID
	var groupIDs []string
	var err error
	if len(args) == 0 {
		userID, serviceAccountID, _, err := common.GetMember(ctx, c.Rig, c.Prompter)
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

		_, groupID, err = c.Prompter.Select("Select group", gIDs)
		if err != nil {
			return err
		}
	} else {
		_, groupID, err = common.GetGroup(ctx, groupID, c.Rig, c.Prompter)
		if err != nil {
			return err
		}
	}

	req := &group.AddMemberRequest{
		GroupId: groupID,
		MemberIds: []*group.MemberID{
			memberID,
		},
	}

	_, err = c.Rig.Group().AddMember(ctx, &connect.Request[group.AddMemberRequest]{
		Msg: req,
	})
	if err != nil {
		return err
	}

	cmd.Printf("User added to group %s\n", gname)
	return nil
}
