package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func UserDelete(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := utils.GetUser(ctx, identifier, nc)
	if err != nil {
		return err
	}

	_, err = nc.User().Delete(ctx, connect.NewRequest(&user.DeleteRequest{
		UserId: id,
	}))
	if err != nil {
		return err
	}

	cmd.Printf("User deleted\n")
	return nil
}
