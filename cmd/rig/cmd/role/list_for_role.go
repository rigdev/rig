package role

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/role"
	"github.com/rigdev/rig-go-api/model"
	"github.com/spf13/cobra"
)

func (c *Cmd) listGroupsForRole(ctx context.Context, cmd *cobra.Command, args []string) error {
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

		_, roleID, err = c.Prompter.Select("Select role to list groups for", roleIDs)
		if err != nil {
			return err
		}
	}

	groupsResp, err := c.Rig.Role().ListAssignees(ctx, connect.NewRequest(&role.ListAssigneesRequest{
		RoleId: roleID,
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
	}))
	if err != nil {
		return err
	}

	if len(groupsResp.Msg.GetEntityIds()) == 0 {
		cmd.Println("No groups assigned to role")
		return nil
	}

	for _, ID := range groupsResp.Msg.GetEntityIds() {
		cmd.Println("Group: ", ID)
	}

	return nil
}
