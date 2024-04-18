package role

import (
	"context"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/role"
	"github.com/rigdev/rig-go-api/model"
	"github.com/spf13/cobra"
)

// ListForGroup lists all roles for a group
func (c *Cmd) listRolesForGroup(ctx context.Context, cmd *cobra.Command, args []string) error {
	groupID := ""
	if len(args) > 0 {
		groupID = args[0]
	} else {
		groupRes, err := c.Rig.Group().List(ctx, connect.NewRequest(&group.ListRequest{}))
		if err != nil {
			return err
		}

		var groupIDs []string
		for _, g := range groupRes.Msg.GetGroups() {
			groupIDs = append(groupIDs, g.GetGroupId())
		}

		_, groupID, err = c.Prompter.Select("Select group to list roles for", groupIDs)
		if err != nil {
			return err
		}
	}

	rolesResp, err := c.Rig.Role().ListForEntity(ctx, connect.NewRequest(&role.ListForEntityRequest{
		EntityId: &role.EntityID{
			Kind: &role.EntityID_GroupId{
				GroupId: groupID,
			},
		},
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
	}))
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Roles", "ID", "Project", "Environment", "Role Type"})
	for i, roleID := range rolesResp.Msg.GetRoleIds() {
		roleResp, err := c.Rig.Role().Get(ctx, connect.NewRequest(&role.GetRequest{
			RoleId: roleID,
		}))
		if err != nil {
			return err
		}

		r := roleResp.Msg.GetRole()

		project := "*"
		environment := "*"
		if r.GetPermissions() != nil && len(r.GetPermissions()) > 0 {
			project = r.GetPermissions()[0].GetScope().GetProject()
			environment = r.GetPermissions()[0].GetScope().GetEnvironment()
		}

		roleType, ok := r.Metadata["roleType"]
		if !ok {
			roleType = []byte(r.GetRoleId())
		}

		t.AppendRow(table.Row{i + 1, r.GetRoleId(), project, environment, string(roleType)})
	}
	cmd.Println(t.Render())
	return nil
}
