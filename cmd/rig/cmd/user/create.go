package user

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, _ []string) error {
	updates, err := common.GetUserAndPasswordUpdates(username, email, phoneNumber, password)
	if err != nil {
		return err
	}
	res, err := c.Rig.User().Create(ctx, &connect.Request[user.CreateRequest]{
		Msg: &user.CreateRequest{
			Initializers: updates,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("User created with ID:", res.Msg.GetUser().GetUserId())

	_, roleGroup, err := common.PromptSelect("What is the role of the user?",
		[]string{"admin", "owner", "developer", "viewer"})
	if err != nil {
		return err
	}

	_, err = c.Rig.Group().AddMember(ctx, &connect.Request[group.AddMemberRequest]{
		Msg: &group.AddMemberRequest{
			GroupId: roleGroup,
			MemberIds: []*group.MemberID{
				{
					Kind: &group.MemberID_UserId{
						UserId: res.Msg.GetUser().GetUserId(),
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
