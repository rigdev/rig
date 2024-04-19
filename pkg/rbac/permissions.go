package rbac

import "github.com/rigdev/rig-go-api/api/v1/role"

// Viewers can view all resources
func GetViewerPermissions(projectID, environmentID string) []*role.Permission {
	return []*role.Permission{
		{
			Action: ActionImageView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceImage),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionImageView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionGroupView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceGroup),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionGroupView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceUser),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionGroupView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceServiceAccount),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionProjectView,
			Scope: &role.Scope{
				Resource:    WithID(ResourceProject, projectID),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionRoleView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceRole),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionServiceAccountView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceServiceAccount),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionUserView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceUser),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionClusterConfigView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCluster),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionSettingsView,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceSettings),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionEnvironmentView,
			Scope: &role.Scope{
				Resource:    WithID(ResourceEnvironment, environmentID),
				Environment: environmentID,
				Project:     projectID,
			},
		},
	}
}

// Developers can do everything a viewer can do,
// plus work with capsules and images in all aspects except creating and deleting them
func GetDeveloperPermissions(projectID, environmentID string) []*role.Permission {
	permissions := GetViewerPermissions(projectID, environmentID)
	return append(permissions, []*role.Permission{
		{
			Action: ActionCapsuleRestartInstance,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeploy,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployAutoscale,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployImage,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployConfigFiles,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployContainer,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployNetwork,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployReplica,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployRollback,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployChron,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployEnvironmentSources,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployEnvironmentVariables,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleDeployServiceAccount,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceServiceAccount),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleAbortRollout,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleExecute,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionImageAdd,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionImageAdd,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceImage),
				Project:     projectID,
				Environment: environmentID,
			},
		},
	}...)
}

// Owners can do everything a developer can do, plus:
// - Capsule Edit. This means Create, Delete and Update capsules
// - Image Delete. This means Delete images
// - Capsule Stop Rollout. This means Stop rollouts
func GetOwnerPermissions(projectID, environmentID string) []*role.Permission {
	permissions := GetDeveloperPermissions(projectID, environmentID)
	return append(permissions, []*role.Permission{
		{
			Action: ActionCapsuleEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionImageDelete,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionImageDelete,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceImage),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionCapsuleStopRollout,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: environmentID,
				Project:     projectID,
			},
		},
	}...)
}

// Admins can do everything - Thing that are not only in the others are:
// - Project Edit. This means Create, Delete and Update projects
// - Settings Edit. This means Update project and user settings
// - Environment Edit. This means Create, Delete and Update environments
// - Role Edit. This means Create, Delete and Update roles
// - Role Assign. Assign roles to users, groups and service accounts
// - Role Revoke. Revoke roles from users, groups and service accounts
// - User Edit. This means Create, Delete and Update users
// - Group Edit. This means Create, Delete and Update groups, aswell as add users to groups
// - Service Account Edit. This means Create, Delete and Update service accounts
func GetAdminPermissions(projectID, environmentID string) []*role.Permission {
	permissions := GetOwnerPermissions(projectID, environmentID)
	return append(permissions, []*role.Permission{
		{
			Action: ActionProjectEdit,
			Scope: &role.Scope{
				Resource:    WithID(ResourceProject, projectID),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionSettingsEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceSettings),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionEnvironmentEdit,
			Scope: &role.Scope{
				Resource:    WithID(ResourceEnvironment, environmentID),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionRoleEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceRole),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionRoleAssign,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceRole),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionRoleRevoke,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceRole),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionUserEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceUser),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionGroupEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceGroup),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionGroupEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceUser),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionGroupEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceServiceAccount),
				Environment: environmentID,
				Project:     projectID,
			},
		},
		{
			Action: ActionServiceAccountEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceServiceAccount),
				Environment: environmentID,
				Project:     projectID,
			},
		},
	}...)
}
