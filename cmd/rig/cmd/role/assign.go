package role

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/role"
	"github.com/spf13/cobra"
)

func (c *Cmd) assign(ctx context.Context, cmd *cobra.Command, args []string) error {
	var roleID string
	var groupID string
	// Get role ID
	if len(args) == 0 {
		resp, err := c.Rig.Role().List(ctx, connect.NewRequest(&role.ListRequest{}))
		if err != nil {
			return err
		}

		var roleIDs []string
		for _, r := range resp.Msg.GetRoles() {
			roleIDs = append(roleIDs, r.GetRoleId())
		}

		_, roleID, err = c.Prompter.Select("Select role:", roleIDs)
		if err != nil {
			return err
		}
	} else {
		roleID = args[0]
	}

	// Get group ID
	if len(args) < 2 {
		resp, err := c.Rig.Group().List(ctx, connect.NewRequest(&group.ListRequest{}))
		if err != nil {
			return err
		}

		var groupIDs []string
		for _, g := range resp.Msg.GetGroups() {
			groupIDs = append(groupIDs, g.GetGroupId())
		}

		_, groupID, err = c.Prompter.Select("Select group:", groupIDs)
		if err != nil {
			return err
		}
	} else {
		groupID = args[1]
	}

	if _, err := c.Rig.Role().Assign(ctx, connect.NewRequest(&role.AssignRequest{
		RoleId: roleID,
		EntityId: &role.EntityID{
			Kind: &role.EntityID_GroupId{
				GroupId: groupID,
			},
		},
	})); err != nil {
		return err
	}

	cmd.Println("Assigned role:", roleID, "to group:", groupID)

	return nil
}
