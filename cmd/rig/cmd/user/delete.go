package user

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) delete(ctx context.Context, cmd *cobra.Command, args []string) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := common.GetUser(ctx, identifier, c.Rig, c.Prompter)
	if err != nil {
		return err
	}

	_, err = c.Rig.User().Delete(ctx, connect.NewRequest(&user.DeleteRequest{
		UserId: id,
	}))
	if err != nil {
		return err
	}

	cmd.Printf("User deleted\n")
	return nil
}
