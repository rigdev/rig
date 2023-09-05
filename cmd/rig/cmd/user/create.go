package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func UserCreate(ctx context.Context, cmd *cobra.Command, args []string, client rig.Client) error {
	updates, err := common.GetUserAndPasswordUpdates(username, email, phoneNumber, password)
	if err != nil {
		return err
	}
	res, err := client.User().Create(ctx, &connect.Request[user.CreateRequest]{
		Msg: &user.CreateRequest{
			Initializers: updates,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("User created with ID:", res.Msg.GetUser().GetUserId())

	return nil
}
