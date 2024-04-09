package user

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, _ []string) error {
	updates, err := c.Prompter.GetUserAndPasswordUpdates(username, email, password)
	if err != nil {
		return err
	}

	if role == "" {
		_, role, err = c.Prompter.Select("What is the role of the user?",
			[]string{"admin", "owner", "developer", "viewer"})
		if err != nil {
			return err
		}
	}

	res, err := c.Rig.User().Create(ctx, &connect.Request[user.CreateRequest]{
		Msg: &user.CreateRequest{
			Initializers:   updates,
			InitialGroupId: role,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("Succesfully created user with ID:", res.Msg.GetUser().GetUserId())

	return nil
}
