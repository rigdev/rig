package role

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/role"
	"github.com/spf13/cobra"
)

func (c *Cmd) delete(ctx context.Context, cmd *cobra.Command, args []string) error {
	roleID := ""
	if len(args) > 0 {
		roleID = args[0]
	} else {
		resp, err := c.Rig.Role().List(ctx, connect.NewRequest(&role.ListRequest{}))
		if err != nil {
			return err
		}

		var roleIDs []string
		for _, r := range resp.Msg.GetRoles() {
			roleIDs = append(roleIDs, r.GetRoleId())
		}

		_, roleID, err = c.Prompter.Select("Select role to delete", roleIDs)
		if err != nil {
			return err
		}
	}

	if _, err := c.Rig.Role().Delete(ctx, connect.NewRequest(&role.DeleteRequest{
		RoleId: roleID,
	})); err != nil {
		return err
	}

	cmd.Println("Deleted role:", roleID)
	return nil
}
