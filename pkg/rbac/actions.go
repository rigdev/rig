package rbac

import (
	"github.com/rigdev/rig-go-api/api/v1/activity/activityconnect"
	"github.com/rigdev/rig-go-api/api/v1/capsule/capsuleconnect"
	"github.com/rigdev/rig-go-api/api/v1/cluster/clusterconnect"
	"github.com/rigdev/rig-go-api/api/v1/environment/environmentconnect"
	"github.com/rigdev/rig-go-api/api/v1/group/groupconnect"
	"github.com/rigdev/rig-go-api/api/v1/image/imageconnect"
	"github.com/rigdev/rig-go-api/api/v1/issue/issueconnect"
	"github.com/rigdev/rig-go-api/api/v1/metrics/metricsconnect"
	"github.com/rigdev/rig-go-api/api/v1/project/projectconnect"
	"github.com/rigdev/rig-go-api/api/v1/role/roleconnect"
	"github.com/rigdev/rig-go-api/api/v1/service_account/service_accountconnect"
	"github.com/rigdev/rig-go-api/api/v1/settings/settingsconnect"
	"github.com/rigdev/rig-go-api/api/v1/user/userconnect"
)

// Settings actions
const (
	ActionSettingsView = "settings:view"
	ActionSettingsEdit = "settings:edit"
)

// Users actions
const (
	ActionUserView = "user:view"
	ActionUserEdit = "user:edit"
)

// Groups actions
const (
	ActionGroupView = "group:view"
	ActionGroupEdit = "group:edit"
)

// Roles actions
const (
	ActionRoleView   = "role:view"
	ActionRoleEdit   = "role:edit"
	ActionRoleAssign = "role:assign"
	ActionRoleRevoke = "role:retract"
)

// Service accounts actions
const (
	ActionServiceAccountView = "serviceaccount:view"
	ActionServiceAccountEdit = "serviceaccount:edit"
)

// Projects actions
const (
	ActionProjectView = "project:view"
	ActionProjectEdit = "project:edit"
)

// Cluster config actions
const (
	ActionClusterConfigView = "clusterconfig:view"
)

// Capsules actions
const (
	// Get, List, GetRollout, ListRollout, ListEvents, Logs, ListImages,
	// ListInstances, ListInstanceStatuses, GetInstanceSatatus, CapsuleMetrics
	ActionCapsuleView = "capsule:view"
	// Update Capsules
	ActionCapsuleEdit = "capsule:edit"
	// Create
	ActionCapsuleCreate = "capsule:create"
	// Delete
	ActionCapsuleDelete = "capsule:delete"
	// Execute
	ActionCapsuleExecute = "capsule:execute"
	// Restart instance
	ActionCapsuleRestartInstance = "capsule:restartinstance"
	// Abort rollout
	ActionCapsuleAbortRollout = "capsule:abortrollout"
	// Stop rollout
	ActionCapsuleStopRollout = "capsule:stoprollout"
	ActionCapsuleGetRevision = "capsule:getRevision"

	// Deploy
	ActionCapsuleDeploy = "capsule:deploy"
	// Horizontally scaling - replicas
	ActionCapsuleDeployReplica = "capsule:deploy:replica"
	// Autoscaling - min, max and cpu threshold - Horizontal scaling
	ActionCapsuleDeployAutoscale = "capsule:deploy:autoscale"
	// Container Settings - Vertical scaling, Environment variables, Command and Args
	ActionCapsuleDeployContainer = "capsule:deploy:container"
	// Config Files - Add and remove config files
	ActionCapsuleDeployConfigFiles = "capsule:deploy:configfiles"
	// Network - Add, remove and update networks
	ActionCapsuleDeployNetwork = "capsule:deploy:network"
	// Rollback
	ActionCapsuleDeployRollback = "capsule:deploy:rollback"
	// Deploy a new image
	ActionCapsuleDeployImage = "capsule:deploy:image"
	// Auto add service account to the capsule
	ActionCapsuleDeployServiceAccount = "capsule:deploy:serviceaccount"
	// Chron jobs
	ActionCapsuleDeployChron = "capsule:deploy:chron"
	// Environment variables - set or remove environment variables
	ActionCapsuleDeployEnvironmentVariables = "capsule:deploy:environmentvariables"
	// Environment sources - set or remove environment sources
	ActionCapsuleDeployEnvironmentSources = "capsule:deploy:environmentsources"
	// Container Settings - Vertical scaling, Environment variables, Command and Args
	ActionCapsuleDeployAnnotations = "capsule:deploy:annotations"
	// Port forwarding
	ActionCapsulePortForward = "capsule:portfoward"
)

// Image actions
const (
	// Create images
	ActionImageAdd = "image:add"
	// Delete images
	ActionImageDelete = "image:delete"
	// View all parts of the build - GetImage, GetImageImageInfo, GetRepositoryInfo, GetImageLogs, GetImageStatus
	ActionImageView = "image:view"
)

// Environment actions
const (
	ActionEnvironmentEdit = "environment:edit"
	// Create and delete ephemeral environments
	ActionEnvironmentEditEphemeral = "environment:edit:ephemeral"
	ActionEnvironmentView          = "environment:view"
)

// Metrics actions
const (
	ActionMetricsView = "metrics:view"
)

// Activity actions
const (
	ActionActivityView = "activity:view"
)

// Issue actions
const (
	ActionIssueView = "issue:view"
)

var CapsuleActionMap = map[string]string{
	capsuleconnect.ServiceWatchStatusProcedure:              ActionCapsuleView,
	capsuleconnect.ServiceGetStatusProcedure:                ActionCapsuleView,
	capsuleconnect.ServiceGetProcedure:                      ActionCapsuleView,
	capsuleconnect.ServiceListProcedure:                     ActionCapsuleView,
	capsuleconnect.ServiceGetRolloutProcedure:               ActionCapsuleView,
	capsuleconnect.ServiceListRolloutsProcedure:             ActionCapsuleView,
	capsuleconnect.ServiceListEventsProcedure:               ActionCapsuleView,
	capsuleconnect.ServiceLogsProcedure:                     ActionCapsuleView,
	capsuleconnect.ServiceListInstancesProcedure:            ActionCapsuleView,
	capsuleconnect.ServiceListInstanceStatusesProcedure:     ActionCapsuleView,
	capsuleconnect.ServiceGetInstanceStatusProcedure:        ActionCapsuleView,
	capsuleconnect.ServiceCapsuleMetricsProcedure:           ActionCapsuleView,
	capsuleconnect.ServiceGetCustomInstanceMetricsProcedure: ActionCapsuleView,
	capsuleconnect.ServiceCreateProcedure:                   ActionCapsuleCreate,
	capsuleconnect.ServiceDeleteProcedure:                   ActionCapsuleDelete,
	capsuleconnect.ServiceUpdateProcedure:                   ActionCapsuleEdit,
	capsuleconnect.ServiceGetJobExecutionsProcedure:         ActionCapsuleView,
	capsuleconnect.ServiceDeployProcedure:                   ActionCapsuleDeploy,
	capsuleconnect.ServiceDeploySetProcedure:                ActionCapsuleDeploy,
	capsuleconnect.ServiceProposeRolloutProcedure:           ActionCapsuleDeploy,
	capsuleconnect.ServiceProposeSetRolloutProcedure:        ActionCapsuleDeploy,
	capsuleconnect.ServiceAbortRolloutProcedure:             ActionCapsuleAbortRollout,
	capsuleconnect.ServiceStopRolloutProcedure:              ActionCapsuleStopRollout,
	capsuleconnect.ServiceExecuteProcedure:                  ActionCapsuleExecute,
	capsuleconnect.ServiceRestartInstanceProcedure:          ActionCapsuleRestartInstance,
	capsuleconnect.ServiceGetRevisionProcedure:              ActionCapsuleGetRevision,
	capsuleconnect.ServiceGetRolloutOfRevisionsProcedure:    ActionCapsuleView,
	capsuleconnect.ServiceGetEffectiveGitSettingsProcedure:  ActionCapsuleView,
	capsuleconnect.ServiceListProposalsProcedure:            ActionCapsuleView,
	capsuleconnect.ServiceListSetProposalsProcedure:         ActionCapsuleView,
	capsuleconnect.ServicePortForwardProcedure:              ActionCapsulePortForward,
	capsuleconnect.ServiceListPipelineStatusesProcedure:     ActionCapsuleView,
	capsuleconnect.ServiceGetPipelineStatusProcedure:        ActionCapsuleView,
	capsuleconnect.ServiceAbortPipelineProcedure:            ActionCapsuleDeploy,
	capsuleconnect.ServicePromotePipelineProcedure:          ActionCapsuleDeploy,
	capsuleconnect.ServiceStartPipelineProcedure:            ActionCapsuleDeploy,
	capsuleconnect.ServiceGetProposalsEnabledProcedure:      ActionCapsuleView,
}

var UserActionMap = map[string]string{
	userconnect.ServiceCreateProcedure:          ActionUserEdit,
	userconnect.ServiceDeleteProcedure:          ActionUserEdit,
	userconnect.ServiceUpdateProcedure:          ActionUserEdit,
	userconnect.ServiceGetProcedure:             ActionUserView,
	userconnect.ServiceGetByIdentifierProcedure: ActionUserView,
	userconnect.ServiceListProcedure:            ActionUserView,
	userconnect.ServiceListSessionsProcedure:    ActionUserView,
}

var GroupActionMap = map[string]string{
	groupconnect.ServiceAddMemberProcedure:           ActionGroupEdit,
	groupconnect.ServiceCreateProcedure:              ActionGroupEdit,
	groupconnect.ServiceDeleteProcedure:              ActionGroupEdit,
	groupconnect.ServiceRemoveMemberProcedure:        ActionGroupEdit,
	groupconnect.ServiceUpdateProcedure:              ActionGroupEdit,
	groupconnect.ServiceGetProcedure:                 ActionGroupView,
	groupconnect.ServiceListProcedure:                ActionGroupView,
	groupconnect.ServiceListGroupsForMemberProcedure: ActionGroupView,
	groupconnect.ServiceListMembersProcedure:         ActionGroupView,
}

var ServiceAccountActionMap = map[string]string{
	service_accountconnect.ServiceCreateProcedure: ActionServiceAccountEdit,
	service_accountconnect.ServiceDeleteProcedure: ActionServiceAccountEdit,
	service_accountconnect.ServiceListProcedure:   ActionServiceAccountView,
}

var ProjectActionMap = map[string]string{
	projectconnect.ServiceCreateProcedure:                           ActionProjectEdit,
	projectconnect.ServiceDeleteProcedure:                           ActionProjectEdit,
	projectconnect.ServiceGetProcedure:                              ActionProjectView,
	projectconnect.ServiceGetCustomObjectMetricsProcedure:           ActionProjectView,
	projectconnect.ServiceGetObjectsByKindProcedure:                 ActionProjectView,
	projectconnect.ServiceListProcedure:                             ActionProjectView,
	projectconnect.ServicePublicKeyProcedure:                        ActionProjectView,
	projectconnect.ServiceGetEffectiveGitSettingsProcedure:          ActionProjectView,
	projectconnect.ServiceGetEffectivePipelineSettingsProcedure:     ActionProjectView,
	projectconnect.ServiceGetEffectiveNotificationSettingsProcedure: ActionProjectView,
	projectconnect.ServiceUpdateProcedure:                           ActionProjectEdit,
}

var ClusterActionMap = map[string]string{
	clusterconnect.ServiceGetConfigProcedure:    ActionClusterConfigView,
	clusterconnect.ServiceGetConfigsProcedure:   ActionClusterConfigView,
	clusterconnect.ServiceListProcedure:         ActionClusterConfigView,
	clusterconnect.ServiceListNodesProcedure:    ActionClusterConfigView,
	clusterconnect.ServiceListNodePodsProcedure: ActionClusterConfigView,
}

var ImageActionMap = map[string]string{
	imageconnect.ServiceAddProcedure:               ActionImageAdd,
	imageconnect.ServiceDeleteProcedure:            ActionImageDelete,
	imageconnect.ServiceGetProcedure:               ActionImageView,
	imageconnect.ServiceGetImageInfoProcedure:      ActionImageView,
	imageconnect.ServiceGetRepositoryInfoProcedure: ActionImageView,
	imageconnect.ServiceListProcedure:              ActionImageView,
}

var RoleActionMap = map[string]string{
	roleconnect.ServiceCreateProcedure:        ActionRoleEdit,
	roleconnect.ServiceDeleteProcedure:        ActionRoleEdit,
	roleconnect.ServiceUpdateProcedure:        ActionRoleEdit,
	roleconnect.ServiceRevokeProcedure:        ActionRoleRevoke,
	roleconnect.ServiceAssignProcedure:        ActionRoleAssign,
	roleconnect.ServiceGetProcedure:           ActionRoleView,
	roleconnect.ServiceListProcedure:          ActionRoleView,
	roleconnect.ServiceListForEntityProcedure: ActionRoleView,
	roleconnect.ServiceListAssigneesProcedure: ActionRoleView,
}

var SettingsActionMap = map[string]string{
	settingsconnect.ServiceGetSettingsProcedure:       ActionSettingsView,
	settingsconnect.ServiceGetLicenseInfoProcedure:    ActionSettingsView,
	settingsconnect.ServiceGetConfigurationProcedure:  ActionSettingsView,
	settingsconnect.ServiceUpdateSettingsProcedure:    ActionSettingsEdit,
	settingsconnect.ServiceGetGitStoreStatusProcedure: ActionSettingsView,
}

var EnvironmentActionMap = map[string]string{
	environmentconnect.ServiceCreateProcedure:        ActionEnvironmentEdit,
	environmentconnect.ServiceDeleteProcedure:        ActionEnvironmentEdit,
	environmentconnect.ServiceGetNamespacesProcedure: ActionEnvironmentView,
	environmentconnect.ServiceListProcedure:          ActionEnvironmentView,
	environmentconnect.ServiceUpdateProcedure:        ActionEnvironmentEdit,
	environmentconnect.ServiceGetProcedure:           ActionEnvironmentView,
}

var MetricsActionMap = map[string]string{
	metricsconnect.ServiceGetMetricsProcedure:           ActionMetricsView,
	metricsconnect.ServiceGetMetricsManyProcedure:       ActionMetricsView,
	metricsconnect.ServiceGetMetricsExpressionProcedure: ActionMetricsView,
}

var ActivityActionMap = map[string]string{
	activityconnect.ServiceGetActivitiesProcedure: ActionActivityView,
}

var IssueActionMap = map[string]string{
	issueconnect.ServiceGetIssuesProcedure: ActionIssueView,
}
