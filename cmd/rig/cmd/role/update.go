package role

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	environment_api "github.com/rigdev/rig-go-api/api/v1/environment"
	project_api "github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/api/v1/role"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/rbac"
	"github.com/spf13/cobra"
)

func (c *Cmd) update(ctx context.Context, cmd *cobra.Command, args []string) error {
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

		_, roleID, err = c.Prompter.Select("Select role to update", roleIDs)
		if err != nil {
			return err
		}
	}

	roleResp, err := c.Rig.Role().Get(ctx, connect.NewRequest(&role.GetRequest{
		RoleId: roleID,
	}))
	if err != nil {
		return err
	}

	currRollType := ""
	currRollTypeBytes, ok := roleResp.Msg.GetRole().GetMetadata()["roleType"]
	if ok {
		currRollType = string(currRollTypeBytes)
	}

	currProject := roleResp.Msg.GetRole().GetPermissions()[0].GetScope().GetProject()
	currEnv := roleResp.Msg.GetRole().GetPermissions()[0].GetScope().GetEnvironment()

	// Nothing set to update so we prompt
	if updateRoleType == "" && updateProject == "" && updateEnvironment == "" {
		for {
			cmd.Printf("Current Role Config:\nRole Type: %v\nProject: %v\nEnvironment: %v\n", currRollType, currProject, currEnv)
			choices := []string{"Role Type", "Project", "Environment", "Done"}
			_, choice, err := c.Prompter.Select("Select what to update", choices)
			if err != nil {
				return err
			}

			switch choice {
			case "Role Type":
				roleChoices := []string{"admin", "owner", "developer", "viewer"}
				// remove the current role type from choices
				for i, r := range roleChoices {
					if r == currRollType {
						roleChoices = append(roleChoices[:i], roleChoices[i+1:]...)
						break
					}
				}

				_, updateRoleType, err := c.Prompter.Select("Select the role type to update to", roleChoices)
				if err != nil {
					return err
				}
				currRollType = updateRoleType
			case "Project":
				projectResp, err := c.Rig.Project().List(ctx, connect.NewRequest(&project_api.ListRequest{}))
				if err != nil {
					return err
				}

				projectIDsChoices := []string{"*"}
				for _, p := range projectResp.Msg.GetProjects() {
					projectIDsChoices = append(projectIDsChoices, p.GetProjectId())
				}

				// remove the current project from choices
				for i, p := range projectIDsChoices {
					if p == currProject {
						projectIDsChoices = append(projectIDsChoices[:i], projectIDsChoices[i+1:]...)
						break
					}
				}

				_, updateProject, err = c.Prompter.Select("Select the project to update to", projectIDsChoices)
				if err != nil {
					return err
				}
				currProject = updateProject
			case "Environment":
				envResp, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment_api.ListRequest{}))
				if err != nil {
					return err
				}

				envIDsChoices := []string{"*"}
				for _, e := range envResp.Msg.GetEnvironments() {
					envIDsChoices = append(envIDsChoices, e.GetEnvironmentId())
				}

				// remove the current environment from choices
				for i, e := range envIDsChoices {
					if e == currEnv {
						envIDsChoices = append(envIDsChoices[:i], envIDsChoices[i+1:]...)
						break
					}
				}

				_, updateEnvironment, err = c.Prompter.Select("Select the environment to update to", envIDsChoices)
				if err != nil {
					return err
				}
				currEnv = updateEnvironment
			case "Done":
				goto done
			}
		}
	done:
	}

	if updateProject == "" {
		updateProject = currProject
	}
	if updateEnvironment == "" {
		updateEnvironment = currEnv
	}
	if updateRoleType == "" {
		updateRoleType = currRollType
	}

	var permissions []*role.Permission
	switch updateRoleType {
	case "admin":
		permissions = rbac.GetAdminPermissions(updateProject, updateEnvironment)
	case "owner":
		permissions = rbac.GetOwnerPermissions(updateProject, updateEnvironment)
	case "developer":
		permissions = rbac.GetDeveloperPermissions(updateProject, updateEnvironment)
	case "viewer":
		permissions = rbac.GetViewerPermissions(updateProject, updateEnvironment)
	default:
		return fmt.Errorf("invalid role type: %v", updateRoleType)
	}

	assigneesResp, err := c.Rig.Role().ListAssignees(ctx, connect.NewRequest(&role.ListAssigneesRequest{
		RoleId: roleID,
	}))
	if err != nil {
		return err
	}

	if _, err := c.Rig.Role().Delete(ctx, connect.NewRequest(&role.DeleteRequest{
		RoleId: roleID,
	})); err != nil {
		return err
	}

	if _, err := c.Rig.Role().Create(ctx, connect.NewRequest(&role.CreateRequest{
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
						Value: []byte(updateRoleType),
					},
				},
			},
		},
	})); err != nil {
		return err
	}

	for _, a := range assigneesResp.Msg.GetEntityIds() {
		if _, err := c.Rig.Role().Assign(ctx, connect.NewRequest(&role.AssignRequest{
			RoleId: roleID,
			EntityId: &role.EntityID{
				Kind: &role.EntityID_GroupId{
					GroupId: a,
				},
			},
		})); err != nil {
			return err
		}
	}

	cmd.Printf("Role %s updated successfully with role: %s, project: %s and environment: %s\n",
		roleID, updateRoleType, updateProject, updateEnvironment)
	return nil
}
