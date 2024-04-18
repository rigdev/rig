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
				Resource:    WithWildcard(ResourceProject),
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
				Resource:    WithWildcard(ResourceEnvironment),
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
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeploy,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployAutoscale,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployImage,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployConfigFiles,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployContainer,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployNetwork,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployReplica,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployRollback,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployChron,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployEnvironmentSources,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployEnvironmentVariables,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleDeployServiceAccount,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceServiceAccount),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleAbortRollout,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleExecute,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionImageAdd,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionImageAdd,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceImage),
				Project:     "*",
				Environment: "*",
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
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionImageDelete,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionImageDelete,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceImage),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionCapsuleStopRollout,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceCapsule),
				Environment: "*",
				Project:     "*",
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
				Resource:    WithWildcard(ResourceProject),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionSettingsEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceSettings),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionEnvironmentEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceEnvironment),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionRoleEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceRole),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionRoleAssign,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceRole),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionRoleRevoke,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceRole),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionUserEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceUser),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionGroupEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceGroup),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionGroupEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceUser),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionGroupEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceServiceAccount),
				Environment: "*",
				Project:     "*",
			},
		},
		{
			Action: ActionServiceAccountEdit,
			Scope: &role.Scope{
				Resource:    WithWildcard(ResourceServiceAccount),
				Environment: "*",
				Project:     "*",
			},
		},
	}...)
}
