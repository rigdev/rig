package role

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/role"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/rbac"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, args []string) error {
	roleID := ""
	var err error
	if len(args) > 0 {
		roleID = args[0]
	} else {
		roleID, err = c.Prompter.Input("Enter Role ID:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	var permissions []*role.Permission
	switch roleType {
	case "admin":
		permissions = rbac.GetAdminPermissions(project, environment)
	case "owner":
		permissions = rbac.GetOwnerPermissions(project, environment)
	case "developer":
		permissions = rbac.GetDeveloperPermissions(project, environment)
	case "viewer":
		permissions = rbac.GetViewerPermissions(project, environment)
	default:
		return fmt.Errorf("invalid role type: %v", roleType)
	}

	if _, err = c.Rig.Role().Create(ctx, connect.NewRequest(&role.CreateRequest{
		RoleId:      roleID,
		Permissions: permissions,
	})); err != nil {
		return err
	}

	if _, err = c.Rig.Role().Update(ctx, connect.NewRequest(&role.UpdateRequest{
		RoleId: roleID,
		Updates: []*role.Update{
			{
				Update: &role.Update_SetMetadata{
					SetMetadata: &model.Metadata{
						Key:   "roleType",
						Value: []byte(roleType),
					},
				},
			},
		},
	})); err != nil {
		return err
	}

	cmd.Printf("Created role %s of type %s with access to project %s and environment %s \n",
		roleID, roleType, project, environment)
	return nil
}
