<a name="top"></a>







### api.v1.activity.Service
<a name="api-v1-activity-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.activity.Service/GetActivities | [GetActivitiesRequest](#api-v1-activity-GetActivitiesRequest) | [GetActivitiesResponse](#api-v1-activity-GetActivitiesResponse) | Get Activities |










### api.v1.authentication.Service
<a name="api-v1-authentication-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.authentication.Service/Login | [LoginRequest](#api-v1-authentication-LoginRequest) | [LoginResponse](#api-v1-authentication-LoginResponse) | Login authenticats a user and returns a access/refresh token |
| /api.v1.authentication.Service/Logout | [LogoutRequest](#api-v1-authentication-LogoutRequest) | [LogoutResponse](#api-v1-authentication-LogoutResponse) | Logout validates the access token and blocks it afterwards |
| /api.v1.authentication.Service/Get | [GetRequest](#api-v1-authentication-GetRequest) | [GetResponse](#api-v1-authentication-GetResponse) | Get the logged in user |
| /api.v1.authentication.Service/Register | [RegisterRequest](#api-v1-authentication-RegisterRequest) | [RegisterResponse](#api-v1-authentication-RegisterResponse) | Register creates a new user |
| /api.v1.authentication.Service/SendPasswordReset | [SendPasswordResetRequest](#api-v1-authentication-SendPasswordResetRequest) | [SendPasswordResetResponse](#api-v1-authentication-SendPasswordResetResponse) | Send reset password email to the user |
| /api.v1.authentication.Service/ResetPassword | [ResetPasswordRequest](#api-v1-authentication-ResetPasswordRequest) | [ResetPasswordResponse](#api-v1-authentication-ResetPasswordResponse) | Reset password of the user |
| /api.v1.authentication.Service/Delete | [DeleteRequest](#api-v1-authentication-DeleteRequest) | [DeleteResponse](#api-v1-authentication-DeleteResponse) | Delete logged in user |
| /api.v1.authentication.Service/RefreshToken | [RefreshTokenRequest](#api-v1-authentication-RefreshTokenRequest) | [RefreshTokenResponse](#api-v1-authentication-RefreshTokenResponse) | Refresh logged in token pair |
| /api.v1.authentication.Service/GetAuthConfig | [GetAuthConfigRequest](#api-v1-authentication-GetAuthConfigRequest) | [GetAuthConfigResponse](#api-v1-authentication-GetAuthConfigResponse) | Get auth config for how available login methods |
| /api.v1.authentication.Service/VerifyEmail | [VerifyEmailRequest](#api-v1-authentication-VerifyEmailRequest) | [VerifyEmailResponse](#api-v1-authentication-VerifyEmailResponse) | Verify email |
| /api.v1.authentication.Service/VerifyPhoneNumber | [VerifyPhoneNumberRequest](#api-v1-authentication-VerifyPhoneNumberRequest) | [VerifyPhoneNumberResponse](#api-v1-authentication-VerifyPhoneNumberResponse) | Verify phone number |
| /api.v1.authentication.Service/SendVerificationEmail | [SendVerificationEmailRequest](#api-v1-authentication-SendVerificationEmailRequest) | [SendVerificationEmailResponse](#api-v1-authentication-SendVerificationEmailResponse) |  |


























### api.v1.capsule.Service
<a name="api-v1-capsule-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.capsule.Service/Create | [CreateRequest](#api-v1-capsule-CreateRequest) | [CreateResponse](#api-v1-capsule-CreateResponse) | Create a new capsule. |
| /api.v1.capsule.Service/Get | [GetRequest](#api-v1-capsule-GetRequest) | [GetResponse](#api-v1-capsule-GetResponse) | Get a capsule by id. |
| /api.v1.capsule.Service/Delete | [DeleteRequest](#api-v1-capsule-DeleteRequest) | [DeleteResponse](#api-v1-capsule-DeleteResponse) | Delete a capsule. |
| /api.v1.capsule.Service/Logs | [LogsRequest](#api-v1-capsule-LogsRequest) | [LogsResponse](#api-v1-capsule-LogsResponse) stream | Logs returns (and streams) the log output of a capsule. |
| /api.v1.capsule.Service/Update | [UpdateRequest](#api-v1-capsule-UpdateRequest) | [UpdateResponse](#api-v1-capsule-UpdateResponse) | Update a capsule. |
| /api.v1.capsule.Service/List | [ListRequest](#api-v1-capsule-ListRequest) | [ListResponse](#api-v1-capsule-ListResponse) | Lists all capsules for current project. |
| /api.v1.capsule.Service/Deploy | [DeployRequest](#api-v1-capsule-DeployRequest) | [DeployResponse](#api-v1-capsule-DeployResponse) | Deploy changes to a capsule. When deploying, a new rollout will be initiated. Only one rollout can be running at a single point in time. Use `Abort` to abort an already running rollout. |
| /api.v1.capsule.Service/DeploySet | [DeploySetRequest](#api-v1-capsule-DeploySetRequest) | [DeploySetResponse](#api-v1-capsule-DeploySetResponse) |  |
| /api.v1.capsule.Service/ProposeRollout | [ProposeRolloutRequest](#api-v1-capsule-ProposeRolloutRequest) | [ProposeRolloutResponse](#api-v1-capsule-ProposeRolloutResponse) |  |
| /api.v1.capsule.Service/ProposeSetRollout | [ProposeSetRolloutRequest](#api-v1-capsule-ProposeSetRolloutRequest) | [ProposeSetRolloutResponse](#api-v1-capsule-ProposeSetRolloutResponse) |  |
| /api.v1.capsule.Service/ListProposals | [ListProposalsRequest](#api-v1-capsule-ListProposalsRequest) | [ListProposalsResponse](#api-v1-capsule-ListProposalsResponse) |  |
| /api.v1.capsule.Service/ListSetProposals | [ListSetProposalsRequest](#api-v1-capsule-ListSetProposalsRequest) | [ListSetProposalsResponse](#api-v1-capsule-ListSetProposalsResponse) |  |
| /api.v1.capsule.Service/GetProposalsEnabled | [GetProposalsEnabledRequest](#api-v1-capsule-GetProposalsEnabledRequest) | [GetProposalsEnabledResponse](#api-v1-capsule-GetProposalsEnabledResponse) |  |
| /api.v1.capsule.Service/ListInstances | [ListInstancesRequest](#api-v1-capsule-ListInstancesRequest) | [ListInstancesResponse](#api-v1-capsule-ListInstancesResponse) | Lists all instances for the capsule. |
| /api.v1.capsule.Service/RestartInstance | [RestartInstanceRequest](#api-v1-capsule-RestartInstanceRequest) | [RestartInstanceResponse](#api-v1-capsule-RestartInstanceResponse) | Restart a single capsule instance. |
| /api.v1.capsule.Service/GetRollout | [GetRolloutRequest](#api-v1-capsule-GetRolloutRequest) | [GetRolloutResponse](#api-v1-capsule-GetRolloutResponse) | Get a single rollout by ID. |
| /api.v1.capsule.Service/ListRollouts | [ListRolloutsRequest](#api-v1-capsule-ListRolloutsRequest) | [ListRolloutsResponse](#api-v1-capsule-ListRolloutsResponse) | Lists all rollouts for the capsule. |
| /api.v1.capsule.Service/WatchRollouts | [WatchRolloutsRequest](#api-v1-capsule-WatchRolloutsRequest) | [WatchRolloutsResponse](#api-v1-capsule-WatchRolloutsResponse) stream | Stream rollouts for a capsule. |
| /api.v1.capsule.Service/AbortRollout | [AbortRolloutRequest](#api-v1-capsule-AbortRolloutRequest) | [AbortRolloutResponse](#api-v1-capsule-AbortRolloutResponse) | Abort the rollout. |
| /api.v1.capsule.Service/StopRollout | [StopRolloutRequest](#api-v1-capsule-StopRolloutRequest) | [StopRolloutResponse](#api-v1-capsule-StopRolloutResponse) | Stop a Rollout, removing all resources associated with it. |
| /api.v1.capsule.Service/ListEvents | [ListEventsRequest](#api-v1-capsule-ListEventsRequest) | [ListEventsResponse](#api-v1-capsule-ListEventsResponse) | List capsule events. |
| /api.v1.capsule.Service/CapsuleMetrics | [CapsuleMetricsRequest](#api-v1-capsule-CapsuleMetricsRequest) | [CapsuleMetricsResponse](#api-v1-capsule-CapsuleMetricsResponse) | Get metrics for a capsule |
| /api.v1.capsule.Service/GetInstanceStatus | [GetInstanceStatusRequest](#api-v1-capsule-GetInstanceStatusRequest) | [GetInstanceStatusResponse](#api-v1-capsule-GetInstanceStatusResponse) | GetInstanceStatus returns the current status for the given instance. |
| /api.v1.capsule.Service/ListInstanceStatuses | [ListInstanceStatusesRequest](#api-v1-capsule-ListInstanceStatusesRequest) | [ListInstanceStatusesResponse](#api-v1-capsule-ListInstanceStatusesResponse) | ListInstanceStatuses lists the status of all instances. |
| /api.v1.capsule.Service/WatchInstanceStatuses | [WatchInstanceStatusesRequest](#api-v1-capsule-WatchInstanceStatusesRequest) | [WatchInstanceStatusesResponse](#api-v1-capsule-WatchInstanceStatusesResponse) stream | Stream Instance Statuses of a capsule. |
| /api.v1.capsule.Service/Execute | [ExecuteRequest](#api-v1-capsule-ExecuteRequest) stream | [ExecuteResponse](#api-v1-capsule-ExecuteResponse) stream | Execute executes a command in a given in instance, and returns the output along with an exit code. |
| /api.v1.capsule.Service/PortForward | [PortForwardRequest](#api-v1-capsule-PortForwardRequest) stream | [PortForwardResponse](#api-v1-capsule-PortForwardResponse) stream | PortForward establishes a port-forwarding for the port to the given instance. |
| /api.v1.capsule.Service/GetCustomInstanceMetrics | [GetCustomInstanceMetricsRequest](#api-v1-capsule-GetCustomInstanceMetricsRequest) | [GetCustomInstanceMetricsResponse](#api-v1-capsule-GetCustomInstanceMetricsResponse) |  |
| /api.v1.capsule.Service/GetJobExecutions | [GetJobExecutionsRequest](#api-v1-capsule-GetJobExecutionsRequest) | [GetJobExecutionsResponse](#api-v1-capsule-GetJobExecutionsResponse) | Get list of job executions performed by the Capsule. |
| /api.v1.capsule.Service/GetStatus | [GetStatusRequest](#api-v1-capsule-GetStatusRequest) | [GetStatusResponse](#api-v1-capsule-GetStatusResponse) |  |
| /api.v1.capsule.Service/GetRevision | [GetRevisionRequest](#api-v1-capsule-GetRevisionRequest) | [GetRevisionResponse](#api-v1-capsule-GetRevisionResponse) |  |
| /api.v1.capsule.Service/GetRolloutOfRevisions | [GetRolloutOfRevisionsRequest](#api-v1-capsule-GetRolloutOfRevisionsRequest) | [GetRolloutOfRevisionsResponse](#api-v1-capsule-GetRolloutOfRevisionsResponse) |  |
| /api.v1.capsule.Service/WatchStatus | [WatchStatusRequest](#api-v1-capsule-WatchStatusRequest) | [WatchStatusResponse](#api-v1-capsule-WatchStatusResponse) stream | Stream the status of a capsule. |
| /api.v1.capsule.Service/GetEffectiveGitSettings | [GetEffectiveGitSettingsRequest](#api-v1-capsule-GetEffectiveGitSettingsRequest) | [GetEffectiveGitSettingsResponse](#api-v1-capsule-GetEffectiveGitSettingsResponse) |  |
| /api.v1.capsule.Service/StartPipeline | [StartPipelineRequest](#api-v1-capsule-StartPipelineRequest) | [StartPipelineResponse](#api-v1-capsule-StartPipelineResponse) | Will initiate the pipeline, from the initial environment and it's current rollout. |
| /api.v1.capsule.Service/GetPipelineStatus | [GetPipelineStatusRequest](#api-v1-capsule-GetPipelineStatusRequest) | [GetPipelineStatusResponse](#api-v1-capsule-GetPipelineStatusResponse) |  |
| /api.v1.capsule.Service/PromotePipeline | [PromotePipelineRequest](#api-v1-capsule-PromotePipelineRequest) | [PromotePipelineResponse](#api-v1-capsule-PromotePipelineResponse) | Progress the pipeline to the next environment. |
| /api.v1.capsule.Service/AbortPipeline | [AbortPipelineRequest](#api-v1-capsule-AbortPipelineRequest) | [AbortPipelineResponse](#api-v1-capsule-AbortPipelineResponse) | Abort the pipeline execution. This will stop the pipeline from any further promotions. |
| /api.v1.capsule.Service/ListPipelineStatuses | [ListPipelineStatusesRequest](#api-v1-capsule-ListPipelineStatusesRequest) | [ListPipelineStatusesResponse](#api-v1-capsule-ListPipelineStatusesResponse) |  |








### api.v1.cluster.Service
<a name="api-v1-cluster-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.cluster.Service/List | [ListRequest](#api-v1-cluster-ListRequest) | [ListResponse](#api-v1-cluster-ListResponse) |  |
| /api.v1.cluster.Service/GetConfig | [GetConfigRequest](#api-v1-cluster-GetConfigRequest) | [GetConfigResponse](#api-v1-cluster-GetConfigResponse) | GetConfig returns the config for the cluster. |
| /api.v1.cluster.Service/GetConfigs | [GetConfigsRequest](#api-v1-cluster-GetConfigsRequest) | [GetConfigsResponse](#api-v1-cluster-GetConfigsResponse) | GetConfigs returns the configs for all clusters. |








### api.v1.environment.Service
<a name="api-v1-environment-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.environment.Service/Create | [CreateRequest](#api-v1-environment-CreateRequest) | [CreateResponse](#api-v1-environment-CreateResponse) |  |
| /api.v1.environment.Service/Update | [UpdateRequest](#api-v1-environment-UpdateRequest) | [UpdateResponse](#api-v1-environment-UpdateResponse) |  |
| /api.v1.environment.Service/Delete | [DeleteRequest](#api-v1-environment-DeleteRequest) | [DeleteResponse](#api-v1-environment-DeleteResponse) |  |
| /api.v1.environment.Service/List | [ListRequest](#api-v1-environment-ListRequest) | [ListResponse](#api-v1-environment-ListResponse) | List available environments. |
| /api.v1.environment.Service/GetNamespaces | [GetNamespacesRequest](#api-v1-environment-GetNamespacesRequest) | [GetNamespacesResponse](#api-v1-environment-GetNamespacesResponse) |  |
| /api.v1.environment.Service/Get | [GetRequest](#api-v1-environment-GetRequest) | [GetResponse](#api-v1-environment-GetResponse) |  |








### api.v1.group.Service
<a name="api-v1-group-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.group.Service/Create | [CreateRequest](#api-v1-group-CreateRequest) | [CreateResponse](#api-v1-group-CreateResponse) | Create a new group. |
| /api.v1.group.Service/Delete | [DeleteRequest](#api-v1-group-DeleteRequest) | [DeleteResponse](#api-v1-group-DeleteResponse) | Delete a group. |
| /api.v1.group.Service/List | [ListRequest](#api-v1-group-ListRequest) | [ListResponse](#api-v1-group-ListResponse) | List groups. |
| /api.v1.group.Service/Update | [UpdateRequest](#api-v1-group-UpdateRequest) | [UpdateResponse](#api-v1-group-UpdateResponse) | Update group. |
| /api.v1.group.Service/Get | [GetRequest](#api-v1-group-GetRequest) | [GetResponse](#api-v1-group-GetResponse) | Get group. |
| /api.v1.group.Service/AddMember | [AddMemberRequest](#api-v1-group-AddMemberRequest) | [AddMemberResponse](#api-v1-group-AddMemberResponse) | Add a member to a group. |
| /api.v1.group.Service/RemoveMember | [RemoveMemberRequest](#api-v1-group-RemoveMemberRequest) | [RemoveMemberResponse](#api-v1-group-RemoveMemberResponse) | Remove member from group. |
| /api.v1.group.Service/ListMembers | [ListMembersRequest](#api-v1-group-ListMembersRequest) | [ListMembersResponse](#api-v1-group-ListMembersResponse) | Get Group Members. |
| /api.v1.group.Service/ListGroupsForMember | [ListGroupsForMemberRequest](#api-v1-group-ListGroupsForMemberRequest) | [ListGroupsForMemberResponse](#api-v1-group-ListGroupsForMemberResponse) | Get Groups. |






### api.v1.image.Service
<a name="api-v1-image-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.image.Service/GetImageInfo | [GetImageInfoRequest](#api-v1-image-GetImageInfoRequest) | [GetImageInfoResponse](#api-v1-image-GetImageInfoResponse) | Get Information about an image in a image. |
| /api.v1.image.Service/GetRepositoryInfo | [GetRepositoryInfoRequest](#api-v1-image-GetRepositoryInfoRequest) | [GetRepositoryInfoResponse](#api-v1-image-GetRepositoryInfoResponse) | Get Information about a docker registry repository. |
| /api.v1.image.Service/Get | [GetRequest](#api-v1-image-GetRequest) | [GetResponse](#api-v1-image-GetResponse) | Get a image. |
| /api.v1.image.Service/Add | [AddRequest](#api-v1-image-AddRequest) | [AddResponse](#api-v1-image-AddResponse) | Add a new image. Images are immutable and cannot change. Add a new image to make changes from an existing one. |
| /api.v1.image.Service/List | [ListRequest](#api-v1-image-ListRequest) | [ListResponse](#api-v1-image-ListResponse) | List images for a capsule. |
| /api.v1.image.Service/Delete | [DeleteRequest](#api-v1-image-DeleteRequest) | [DeleteResponse](#api-v1-image-DeleteResponse) | Delete a image. |







### api.v1.metrics.Service
<a name="api-v1-metrics-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.metrics.Service/GetMetrics | [GetMetricsRequest](#api-v1-metrics-GetMetricsRequest) | [GetMetricsResponse](#api-v1-metrics-GetMetricsResponse) | Retrieve metrics. metric_type is mandatory, while the rest of the fields in the tags are optional. If project, env or capsule is not specified, they will be treated as wildcards. |
| /api.v1.metrics.Service/GetMetricsMany | [GetMetricsManyRequest](#api-v1-metrics-GetMetricsManyRequest) | [GetMetricsManyResponse](#api-v1-metrics-GetMetricsManyResponse) | Retrive metrics for multiple sets of tags at a time. Metrics within the same set of tags will be in ascending order of timestamp. |
| /api.v1.metrics.Service/GetMetricsExpression | [GetMetricsExpressionRequest](#api-v1-metrics-GetMetricsExpressionRequest) | [GetMetricsExpressionResponse](#api-v1-metrics-GetMetricsExpressionResponse) |  |









### api.v1.project.Service
<a name="api-v1-project-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.project.Service/Create | [CreateRequest](#api-v1-project-CreateRequest) | [CreateResponse](#api-v1-project-CreateResponse) | Create project. |
| /api.v1.project.Service/Delete | [DeleteRequest](#api-v1-project-DeleteRequest) | [DeleteResponse](#api-v1-project-DeleteResponse) | Delete project. |
| /api.v1.project.Service/Get | [GetRequest](#api-v1-project-GetRequest) | [GetResponse](#api-v1-project-GetResponse) | Get project. |
| /api.v1.project.Service/List | [ListRequest](#api-v1-project-ListRequest) | [ListResponse](#api-v1-project-ListResponse) | Get project list. |
| /api.v1.project.Service/Update | [UpdateRequest](#api-v1-project-UpdateRequest) | [UpdateResponse](#api-v1-project-UpdateResponse) | Update updates the profile of the project. |
| /api.v1.project.Service/PublicKey | [PublicKeyRequest](#api-v1-project-PublicKeyRequest) | [PublicKeyResponse](#api-v1-project-PublicKeyResponse) | Get public key. |
| /api.v1.project.Service/GetObjectsByKind | [GetObjectsByKindRequest](#api-v1-project-GetObjectsByKindRequest) | [GetObjectsByKindResponse](#api-v1-project-GetObjectsByKindResponse) | Returns all objects of a given kind. |
| /api.v1.project.Service/GetCustomObjectMetrics | [GetCustomObjectMetricsRequest](#api-v1-project-GetCustomObjectMetricsRequest) | [GetCustomObjectMetricsResponse](#api-v1-project-GetCustomObjectMetricsResponse) | Returns all metrics of a given custom object. |
| /api.v1.project.Service/GetEffectiveGitSettings | [GetEffectiveGitSettingsRequest](#api-v1-project-GetEffectiveGitSettingsRequest) | [GetEffectiveGitSettingsResponse](#api-v1-project-GetEffectiveGitSettingsResponse) |  |
| /api.v1.project.Service/GetEffectivePipelineSettings | [GetEffectivePipelineSettingsRequest](#api-v1-project-GetEffectivePipelineSettingsRequest) | [GetEffectivePipelineSettingsResponse](#api-v1-project-GetEffectivePipelineSettingsResponse) |  |
| /api.v1.project.Service/GetEffectiveNotificationSettings | [GetEffectiveNotificationSettingsRequest](#api-v1-project-GetEffectiveNotificationSettingsRequest) | [GetEffectiveNotificationSettingsResponse](#api-v1-project-GetEffectiveNotificationSettingsResponse) |  |







### api.v1.project.settings.Service
<a name="api-v1-project-settings-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.project.settings.Service/GetSettings | [GetSettingsRequest](#api-v1-project-settings-GetSettingsRequest) | [GetSettingsResponse](#api-v1-project-settings-GetSettingsResponse) | Gets the users settings for the current project. |
| /api.v1.project.settings.Service/UpdateSettings | [UpdateSettingsRequest](#api-v1-project-settings-UpdateSettingsRequest) | [UpdateSettingsResponse](#api-v1-project-settings-UpdateSettingsResponse) | Sets the users settings for the current project. |
| /api.v1.project.settings.Service/GetLicenseInfo | [GetLicenseInfoRequest](#api-v1-project-settings-GetLicenseInfoRequest) | [GetLicenseInfoResponse](#api-v1-project-settings-GetLicenseInfoResponse) | Get License Information. |







### api.v1.role.Service
<a name="api-v1-role-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.role.Service/Create | [CreateRequest](#api-v1-role-CreateRequest) | [CreateResponse](#api-v1-role-CreateResponse) | Create a new role. |
| /api.v1.role.Service/Delete | [DeleteRequest](#api-v1-role-DeleteRequest) | [DeleteResponse](#api-v1-role-DeleteResponse) | Delete role. |
| /api.v1.role.Service/List | [ListRequest](#api-v1-role-ListRequest) | [ListResponse](#api-v1-role-ListResponse) | List roles. |
| /api.v1.role.Service/Update | [UpdateRequest](#api-v1-role-UpdateRequest) | [UpdateResponse](#api-v1-role-UpdateResponse) | Update role |
| /api.v1.role.Service/Get | [GetRequest](#api-v1-role-GetRequest) | [GetResponse](#api-v1-role-GetResponse) | Get role. |
| /api.v1.role.Service/Assign | [AssignRequest](#api-v1-role-AssignRequest) | [AssignResponse](#api-v1-role-AssignResponse) | Assign a role. |
| /api.v1.role.Service/Revoke | [RevokeRequest](#api-v1-role-RevokeRequest) | [RevokeResponse](#api-v1-role-RevokeResponse) | Revoke a role. |
| /api.v1.role.Service/ListForEntity | [ListForEntityRequest](#api-v1-role-ListForEntityRequest) | [ListForEntityResponse](#api-v1-role-ListForEntityResponse) | List roles for an entity. |
| /api.v1.role.Service/ListAssignees | [ListAssigneesRequest](#api-v1-role-ListAssigneesRequest) | [ListAssigneesResponse](#api-v1-role-ListAssigneesResponse) | List Assignees. |







### api.v1.service_account.Service
<a name="api-v1-service_account-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.service_account.Service/Create | [CreateRequest](#api-v1-service_account-CreateRequest) | [CreateResponse](#api-v1-service_account-CreateResponse) | Create a new Service Account. The returned client_id and client_secret can be used as login credentials. Note that the client_secret can only be read out once, at creation. |
| /api.v1.service_account.Service/List | [ListRequest](#api-v1-service_account-ListRequest) | [ListResponse](#api-v1-service_account-ListResponse) | List all service accounts. |
| /api.v1.service_account.Service/Delete | [DeleteRequest](#api-v1-service_account-DeleteRequest) | [DeleteResponse](#api-v1-service_account-DeleteResponse) | Delete a service account. It can take up to the TTL of access tokens for existing sessions using this service_account, to expire. |









### api.v1.settings.Service
<a name="api-v1-settings-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.settings.Service/GetConfiguration | [GetConfigurationRequest](#api-v1-settings-GetConfigurationRequest) | [GetConfigurationResponse](#api-v1-settings-GetConfigurationResponse) |  |
| /api.v1.settings.Service/GetSettings | [GetSettingsRequest](#api-v1-settings-GetSettingsRequest) | [GetSettingsResponse](#api-v1-settings-GetSettingsResponse) |  |
| /api.v1.settings.Service/UpdateSettings | [UpdateSettingsRequest](#api-v1-settings-UpdateSettingsRequest) | [UpdateSettingsResponse](#api-v1-settings-UpdateSettingsResponse) |  |
| /api.v1.settings.Service/GetLicenseInfo | [GetLicenseInfoRequest](#api-v1-settings-GetLicenseInfoRequest) | [GetLicenseInfoResponse](#api-v1-settings-GetLicenseInfoResponse) |  |
| /api.v1.settings.Service/GetGitStoreStatus | [GetGitStoreStatusRequest](#api-v1-settings-GetGitStoreStatusRequest) | [GetGitStoreStatusResponse](#api-v1-settings-GetGitStoreStatusResponse) |  |






### api.v1.tunnel.Service
<a name="api-v1-tunnel-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.tunnel.Service/Tunnel | [TunnelRequest](#api-v1-tunnel-TunnelRequest) stream | [TunnelResponse](#api-v1-tunnel-TunnelResponse) stream |  |







### api.v1.user.Service
<a name="api-v1-user-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.user.Service/Create | [CreateRequest](#api-v1-user-CreateRequest) | [CreateResponse](#api-v1-user-CreateResponse) | Create a new user. |
| /api.v1.user.Service/Update | [UpdateRequest](#api-v1-user-UpdateRequest) | [UpdateResponse](#api-v1-user-UpdateResponse) | Update a users profile and info. |
| /api.v1.user.Service/ListSessions | [ListSessionsRequest](#api-v1-user-ListSessionsRequest) | [ListSessionsResponse](#api-v1-user-ListSessionsResponse) | Get the list of active sessions for the given user. |
| /api.v1.user.Service/Get | [GetRequest](#api-v1-user-GetRequest) | [GetResponse](#api-v1-user-GetResponse) | Get a user by user-id. |
| /api.v1.user.Service/GetByIdentifier | [GetByIdentifierRequest](#api-v1-user-GetByIdentifierRequest) | [GetByIdentifierResponse](#api-v1-user-GetByIdentifierResponse) | Lookup a user by a unique identifier - email, username, phone number etc. |
| /api.v1.user.Service/List | [ListRequest](#api-v1-user-ListRequest) | [ListResponse](#api-v1-user-ListResponse) | List users. |
| /api.v1.user.Service/Delete | [DeleteRequest](#api-v1-user-DeleteRequest) | [DeleteResponse](#api-v1-user-DeleteResponse) | Delete a specific user. |








### api.v1.user.settings.Service
<a name="api-v1-user-settings-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.user.settings.Service/GetSettings | [GetSettingsRequest](#api-v1-user-settings-GetSettingsRequest) | [GetSettingsResponse](#api-v1-user-settings-GetSettingsResponse) | Gets the users settings for the current project. |
| /api.v1.user.settings.Service/UpdateSettings | [UpdateSettingsRequest](#api-v1-user-settings-UpdateSettingsRequest) | [UpdateSettingsResponse](#api-v1-user-settings-UpdateSettingsResponse) | Sets the users settings for the current project. |










<a name="api_v1_capsule_rollout_status-proto"></a>

## api/v1/capsule/rollout/status.proto



<a name="api-v1-capsule-rollout-ConfigureCapsuleStep"></a>

### ConfigureCapsuleStep
A step configuring a capsule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-rollout-StepInfo) |  | Step information. |
| state | [ConfigureResult](#api-v1-capsule-rollout-ConfigureResult) |  | The state of the step. |






<a name="api-v1-capsule-rollout-ConfigureCommitStep"></a>

### ConfigureCommitStep
A step committing the changes to git


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-rollout-StepInfo) |  | Step information |
| commit_hash | [string](#string) |  | The hash of the commit containing the changes |
| commit_url | [string](#string) |  | The url to the commit (if known. May be empty) |






<a name="api-v1-capsule-rollout-ConfigureEnvStep"></a>

### ConfigureEnvStep
A step configuring an environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-rollout-StepInfo) |  | Step information. |
| state | [ConfigureResult](#api-v1-capsule-rollout-ConfigureResult) |  | The result of the environment configuration. |
| is_secret | [bool](#bool) |  | Whether the environment is a secret. |






<a name="api-v1-capsule-rollout-ConfigureFileStep"></a>

### ConfigureFileStep
A step configuring a file.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-rollout-StepInfo) |  | Step information. |
| state | [ConfigureResult](#api-v1-capsule-rollout-ConfigureResult) |  | The result of the file configuration. |
| path | [string](#string) |  | The path of the file. |
| is_secret | [bool](#bool) |  | Whether the file is a secret. |






<a name="api-v1-capsule-rollout-ConfigureStage"></a>

### ConfigureStage
The configure stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StageInfo](#api-v1-capsule-rollout-StageInfo) |  | Stage information. |
| steps | [ConfigureStep](#api-v1-capsule-rollout-ConfigureStep) | repeated | The steps of the stage. |






<a name="api-v1-capsule-rollout-ConfigureStep"></a>

### ConfigureStep
A step of the configure stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| generic | [GenericStep](#api-v1-capsule-rollout-GenericStep) |  | A generic step. |
| configure_capsule | [ConfigureCapsuleStep](#api-v1-capsule-rollout-ConfigureCapsuleStep) |  | A step configuring a capsule. |
| configure_file | [ConfigureFileStep](#api-v1-capsule-rollout-ConfigureFileStep) |  | A step configuring a file. |
| configure_env | [ConfigureEnvStep](#api-v1-capsule-rollout-ConfigureEnvStep) |  | A step configuring an environment. |
| commit | [ConfigureCommitStep](#api-v1-capsule-rollout-ConfigureCommitStep) |  | A step for commiting the changes to git. |






<a name="api-v1-capsule-rollout-CreateResourceStep"></a>

### CreateResourceStep
A step creating a resource.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-rollout-StepInfo) |  | Step information. |
| kind | [string](#string) |  | The kind of the resource. |
| name | [string](#string) |  | The name of the resource. |






<a name="api-v1-capsule-rollout-GenericStep"></a>

### GenericStep
A generic step of a stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-rollout-StepInfo) |  | Step information. |






<a name="api-v1-capsule-rollout-InstancesStep"></a>

### InstancesStep
Information on the instances of the rollout.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-rollout-StepInfo) |  | Step information. |
| num_updated | [uint32](#uint32) |  | The number of updated instances. |
| num_ready | [uint32](#uint32) |  | The number of ready instances. |
| num_stuck | [uint32](#uint32) |  | The number of stuck instances. |
| num_wrong_version | [uint32](#uint32) |  | The number of instances with the wrong version. |






<a name="api-v1-capsule-rollout-ResourceCreationStage"></a>

### ResourceCreationStage
The resource creation stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StageInfo](#api-v1-capsule-rollout-StageInfo) |  | Stage information. |
| steps | [ResourceCreationStep](#api-v1-capsule-rollout-ResourceCreationStep) | repeated | The steps of the stage. |






<a name="api-v1-capsule-rollout-ResourceCreationStep"></a>

### ResourceCreationStep
A step of the resource creation stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| generic | [GenericStep](#api-v1-capsule-rollout-GenericStep) |  | A generic step. |
| create_resource | [CreateResourceStep](#api-v1-capsule-rollout-CreateResourceStep) |  | A step creating a resource. |






<a name="api-v1-capsule-rollout-RunningStage"></a>

### RunningStage
The running stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StageInfo](#api-v1-capsule-rollout-StageInfo) |  | Stage information. |
| steps | [RunningStep](#api-v1-capsule-rollout-RunningStep) | repeated | The steps of the stage. |






<a name="api-v1-capsule-rollout-RunningStep"></a>

### RunningStep
A step of the running stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| generic | [GenericStep](#api-v1-capsule-rollout-GenericStep) |  | A generic step. |
| instances | [InstancesStep](#api-v1-capsule-rollout-InstancesStep) |  | A step containing information on the instances of the rollout. |






<a name="api-v1-capsule-rollout-StageInfo"></a>

### StageInfo
Information about a stage of a rollout.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the stage. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The last time the stage was updated. |
| state | [StageState](#api-v1-capsule-rollout-StageState) |  | The current state of the stage. |
| started_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The time the stage started. |






<a name="api-v1-capsule-rollout-Stages"></a>

### Stages
The three stages of a rollout


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| configure | [ConfigureStage](#api-v1-capsule-rollout-ConfigureStage) |  | The configure stage. |
| resource_creation | [ResourceCreationStage](#api-v1-capsule-rollout-ResourceCreationStage) |  | The resource creation stage. |
| running | [RunningStage](#api-v1-capsule-rollout-RunningStage) |  | The running stage. |






<a name="api-v1-capsule-rollout-Status"></a>

### Status
Status is a representation of the current state of a rollout.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rollout_id | [uint64](#uint64) |  | The ID of the rollout. |
| state | [State](#api-v1-capsule-rollout-State) |  | The current state of the rollout. |
| stages | [Stages](#api-v1-capsule-rollout-Stages) |  | The stages of the rollout. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The last time the rollout was updated. |
| result | [Result](#api-v1-capsule-rollout-Result) |  | The result of the rollout. |






<a name="api-v1-capsule-rollout-StepInfo"></a>

### StepInfo
Information about a step of a stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the step. |
| message | [string](#string) |  | Messages in the step. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The last time the step was updated. |
| state | [StepState](#api-v1-capsule-rollout-StepState) |  | The current state of the step. |
| started_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The time the step started. |








<a name="api-v1-capsule-rollout-ConfigureResult"></a>

### ConfigureResult
The result of a configuration step.

| Name | Number | Description |
| ---- | ------ | ----------- |
| CONFIGURE_RESULT_UNSPECIFIED | 0 | The result is unspecified. |
| CONFIGURE_RESULT_CREATED | 1 | The resource is to be created. |
| CONFIGURE_RESULT_UPDATED | 2 | The resource is to be updated. |
| CONFIGURE_RESULT_NO_CHANGE | 3 | The resource has no change. |
| CONFIGURE_RESULT_DELETED | 4 | The resource is to be deleted. |



<a name="api-v1-capsule-rollout-Result"></a>

### Result
Different result of a rollout.

| Name | Number | Description |
| ---- | ------ | ----------- |
| RESULT_UNSPECIFIED | 0 | The result is unspecified. |
| RESULT_REPLACED | 1 | The rollout has been replaced. |
| RESULT_FAILED | 2 | The rollout has failed. |
| RESULT_ABORTED | 3 | The rollout has been aborted. |
| RESULT_ROLLBACK | 4 | The rollout has been rolled back. |



<a name="api-v1-capsule-rollout-StageState"></a>

### StageState
Different states a stage can be in.

| Name | Number | Description |
| ---- | ------ | ----------- |
| STAGE_STATE_UNSPECIFIED | 0 | The state is unspecified. |
| STAGE_STATE_DEPLOYING | 1 | The stage is deploying. |
| STAGE_STATE_RUNNING | 2 | The stage is running. |
| STAGE_STATE_STOPPED | 3 | The stage is stopped. |



<a name="api-v1-capsule-rollout-State"></a>

### State
Different states a rollout can be in.

| Name | Number | Description |
| ---- | ------ | ----------- |
| STATE_UNSPECIFIED | 0 | The state is unspecified. |
| STATE_PREPARING | 1 | The rollout is preparing. |
| STATE_CONFIGURE | 2 | The rollout is configuring. |
| STATE_RESOURCE_CREATION | 3 | The rollout is creating resources. |
| STATE_RUNNING | 4 | The rollout is running. |
| STATE_STOPPED | 5 | The rollout is stopped. |



<a name="api-v1-capsule-rollout-StepState"></a>

### StepState
Different states a step can be in.

| Name | Number | Description |
| ---- | ------ | ----------- |
| STEP_STATE_UNSPECIFIED | 0 | The state is unspecified. |
| STEP_STATE_ONGOING | 1 | The step is ongoing. |
| STEP_STATE_FAILED | 2 | The step failed. |
| STEP_STATE_DONE | 3 | The step is done. |








<a name="model_issue-proto"></a>

## model/issue.proto



<a name="model-Issue"></a>

### Issue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| issue_id | [string](#string) |  |  |
| type | [string](#string) |  |  |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| stale_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| closed_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| reference | [Reference](#model-Reference) |  |  |
| message | [string](#string) |  |  |
| level | [Level](#model-Level) |  |  |
| count | [uint32](#uint32) |  |  |






<a name="model-Reference"></a>

### Reference



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| capsule_id | [string](#string) |  |  |
| environment_id | [string](#string) |  |  |
| rollout_id | [uint64](#uint64) |  |  |
| instance_id | [string](#string) |  |  |








<a name="model-Level"></a>

### Level


| Name | Number | Description |
| ---- | ------ | ----------- |
| LEVEL_UNSPECIFIED | 0 |  |
| LEVEL_INFORMATIVE | 1 |  |
| LEVEL_MINOR | 2 |  |
| LEVEL_MAJOR | 3 |  |
| LEVEL_CRITICAL | 4 |  |








<a name="api_v1_activity_activity-proto"></a>

## api/v1/activity/activity.proto



<a name="api-v1-activity-Activity"></a>

### Activity



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| scope | [Scope](#api-v1-activity-Scope) |  |  |
| message | [Message](#api-v1-activity-Message) |  |  |






<a name="api-v1-activity-Message"></a>

### Message



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rollout | [Message.Rollout](#api-v1-activity-Message-Rollout) |  |  |
| project | [Message.Project](#api-v1-activity-Message-Project) |  |  |
| environment | [Message.Environment](#api-v1-activity-Message-Environment) |  |  |
| capsule | [Message.Capsule](#api-v1-activity-Message-Capsule) |  |  |
| user | [Message.User](#api-v1-activity-Message-User) |  |  |






<a name="api-v1-activity-Message-Capsule"></a>

### Message.Capsule



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  |  |
| deleted | [bool](#bool) |  |  |






<a name="api-v1-activity-Message-Environment"></a>

### Message.Environment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_id | [string](#string) |  |  |
| deleted | [bool](#bool) |  |  |






<a name="api-v1-activity-Message-Issue"></a>

### Message.Issue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| level | [model.Level](#model-Level) |  |  |
| rolloutID | [uint64](#uint64) |  |  |
| message | [string](#string) |  |  |
| resolved | [bool](#bool) |  |  |






<a name="api-v1-activity-Message-Project"></a>

### Message.Project



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| deleted | [bool](#bool) |  |  |






<a name="api-v1-activity-Message-Rollout"></a>

### Message.Rollout



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rollout_id | [uint64](#uint64) |  |  |
| state | [api.v1.capsule.rollout.StepState](#api-v1-capsule-rollout-StepState) |  |  |






<a name="api-v1-activity-Message-User"></a>

### Message.User



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| printable_name | [string](#string) |  |  |
| deleted | [bool](#bool) |  |  |






<a name="api-v1-activity-Scope"></a>

### Scope



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  |  |
| environment | [string](#string) |  |  |
| capsule | [string](#string) |  |  |
| user | [string](#string) |  |  |













<a name="model_common-proto"></a>

## model/common.proto



<a name="model-BcryptHashingConfig"></a>

### BcryptHashingConfig
Bcrypt hashing configuration.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cost | [int32](#int32) |  | The cost of the hashing algorithm. |






<a name="model-BcryptHashingInstance"></a>

### BcryptHashingInstance
Bcrypt hashing instance.






<a name="model-HashingConfig"></a>

### HashingConfig
Hashing configuration.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bcrypt | [BcryptHashingConfig](#model-BcryptHashingConfig) |  | if bcrypt is set, use bcrypt. |
| scrypt | [ScryptHashingConfig](#model-ScryptHashingConfig) |  | if scrypt is set, use scrypt. |






<a name="model-HashingInstance"></a>

### HashingInstance
Hashing instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| config | [HashingConfig](#model-HashingConfig) |  | The hashing configuration. |
| hash | [bytes](#bytes) |  | A hash |
| bcrypt | [BcryptHashingInstance](#model-BcryptHashingInstance) |  | if bcrypt is set, this bcrypt instance was used. |
| scrypt | [ScryptHashingInstance](#model-ScryptHashingInstance) |  | if scrypt is set, this scrypt instance was used. |






<a name="model-Pagination"></a>

### Pagination
Pagination option.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| offset | [uint32](#uint32) |  | Where to start the pagination. |
| limit | [uint32](#uint32) |  | How many items to return. |
| descending | [bool](#bool) |  | Whether to sort in descending order. |






<a name="model-ScryptHashingConfig"></a>

### ScryptHashingConfig
Scrypt hashing configuration.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| signer_key | [string](#string) |  | The key used to sign the salt. |
| salt_separator | [string](#string) |  | The salt separator. |
| rounds | [int32](#int32) |  | The number of rounds in the algorithm. |
| mem_cost | [int32](#int32) |  | The memory cost of the algorithm. |
| p | [int32](#int32) |  | The parallelization factor of the algorithm. |
| key_len | [int32](#int32) |  | The length of the key. |






<a name="model-ScryptHashingInstance"></a>

### ScryptHashingInstance
Scrypt hashing instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| salt | [bytes](#bytes) |  | The salt used to hash the password. |













<a name="api_v1_activity_service-proto"></a>

## api/v1/activity/service.proto



<a name="api-v1-activity-ActivityFilter"></a>

### ActivityFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_filter | [string](#string) |  |  |
| environment_filter | [string](#string) |  |  |
| capsule_filter | [string](#string) |  |  |
| user_identifier_filter | [string](#string) |  |  |






<a name="api-v1-activity-GetActivitiesRequest"></a>

### GetActivitiesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| to | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| pagination | [model.Pagination](#model-Pagination) |  |  |
| filter | [ActivityFilter](#api-v1-activity-ActivityFilter) |  |  |






<a name="api-v1-activity-GetActivitiesResponse"></a>

### GetActivitiesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| activities | [Activity](#api-v1-activity-Activity) | repeated |  |
| total | [uint64](#uint64) |  |  |













<a name="model_auth-proto"></a>

## model/auth.proto



<a name="model-AuthMethod"></a>

### AuthMethod
Message that tells how the user was authenticated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login_type | [LoginType](#model-LoginType) |  |  |








<a name="model-LoginType"></a>

### LoginType
The LoginType enum defines the type of login.

| Name | Number | Description |
| ---- | ------ | ----------- |
| LOGIN_TYPE_UNSPECIFIED | 0 | Default value. |
| LOGIN_TYPE_EMAIL_PASSWORD | 1 | Email and password login. |
| LOGIN_TYPE_PHONE_PASSWORD | 2 | deprecated: text is not supported - Phone number and password login. |
| LOGIN_TYPE_USERNAME_PASSWORD | 3 | Username and password login. |
| LOGIN_TYPE_SSO | 4 | SSO Login |








<a name="model_author-proto"></a>

## model/author.proto



<a name="model-Author"></a>

### Author
Author of a change.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identifier | [string](#string) |  | Cached identifier of the author, e.g. email or username at the time of change. |
| printable_name | [string](#string) |  | Cached pretty-printed name of the author at the time of change. |
| user_id | [string](#string) |  | if a user did the change |
| service_account_id | [string](#string) |  | if a service account did the change |













<a name="model_user-proto"></a>

## model/user.proto



<a name="model-MemberEntry"></a>

### MemberEntry
Entry model of a group member - placed in models to prevent cyclic imports.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [UserEntry](#model-UserEntry) |  | if the member is a user. |
| service_account | [ServiceAccountEntry](#model-ServiceAccountEntry) |  | if the member is a service account. |
| joined_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | when the member joined the group. |






<a name="model-RegisterInfo"></a>

### RegisterInfo
Registering information of a user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| creater_id | [string](#string) |  | Who created the user. |
| method | [RegisterMethod](#model-RegisterMethod) |  | How the user was registered. |






<a name="model-RegisterMethod"></a>

### RegisterMethod
Method used to register a user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| system | [RegisterMethod.System](#model-RegisterMethod-System) |  | system created the user. |
| signup | [RegisterMethod.Signup](#model-RegisterMethod-Signup) |  | user signed up. |






<a name="model-RegisterMethod-Signup"></a>

### RegisterMethod.Signup
if the user was created by signing up.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login_type | [LoginType](#model-LoginType) |  | The login type used to sign up. |






<a name="model-RegisterMethod-System"></a>

### RegisterMethod.System
if the user was created by the system.






<a name="model-ServiceAccountEntry"></a>

### ServiceAccountEntry
Entry model of a service account - placed in models to prevent cyclic
imports.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_account_id | [string](#string) |  | unique id of the service account. |
| name | [string](#string) |  | name of the service account. |
| client_id | [string](#string) |  | client id of the service account. |
| group_ids | [string](#string) | repeated | groups the service account belongs to. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | when the service account was created. |
| created_by | [Author](#model-Author) |  | who created the service account. |






<a name="model-UserEntry"></a>

### UserEntry
Entry model of a user - placed in models to prevent cyclic imports.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  | unique id of the user. |
| printable_name | [string](#string) |  | pretty printable name of a user. |
| register_info | [RegisterInfo](#model-RegisterInfo) |  | how the user was registered. |
| verified | [bool](#bool) |  | whether the user is verified. |
| group_ids | [string](#string) | repeated | groups the user belongs to. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | when the user was created. |






<a name="model-UserIdentifier"></a>

### UserIdentifier
different fields that can identify a user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| username | [string](#string) |  | username is unique. |
| email | [string](#string) |  | email is unique. |
| phone_number | [string](#string) |  | Deprecated: text is not supported - phone number is unique. |






<a name="model-UserInfo"></a>

### UserInfo
Userinfo - placed in models to prevent cyclic imports.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| email | [string](#string) |  | email of the user. |
| username | [string](#string) |  | username of the user. |
| phone_number | [string](#string) |  | Deprecated: text is not supported - phone number of the user. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | when the user was created. |
| group_ids | [string](#string) | repeated | groups the user belongs to. |













<a name="api_v1_authentication_user-proto"></a>

## api/v1/authentication/user.proto



<a name="api-v1-authentication-ClientCredentials"></a>

### ClientCredentials



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_id | [string](#string) |  | ID of the service account |
| client_secret | [string](#string) |  | secret of the service account |






<a name="api-v1-authentication-Token"></a>

### Token



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| access_token | [string](#string) |  | Access token |
| refresh_token | [string](#string) |  | Refresh token |






<a name="api-v1-authentication-UserPassword"></a>

### UserPassword



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identifier | [model.UserIdentifier](#model-UserIdentifier) |  | Identifier of user. This can be username, email etc. |
| password | [string](#string) |  | Password of the user |













<a name="api_v1_authentication_service-proto"></a>

## api/v1/authentication/service.proto



<a name="api-v1-authentication-DeleteRequest"></a>

### DeleteRequest
Request to delete the logged in user. The user ID etc. is taken from the
token.






<a name="api-v1-authentication-DeleteResponse"></a>

### DeleteResponse
Empty response to the delete request.






<a name="api-v1-authentication-GetAuthConfigRequest"></a>

### GetAuthConfigRequest
Empty Request to get the auth config containing the available login
mechanisms and if self-registering is enabled.






<a name="api-v1-authentication-GetAuthConfigResponse"></a>

### GetAuthConfigResponse
Response with the auth config containing the available login mechanisms and
if self-registering is enabled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | The name of the project |
| logo_url | [string](#string) |  | The logo of the project |
| validate_password | [bool](#bool) |  | If to validate password |
| login_types | [model.LoginType](#model-LoginType) | repeated | Array of supported login methods. |
| allows_register | [bool](#bool) |  | True if new users can sign up. |
| sso_options | [SSOOption](#api-v1-authentication-SSOOption) | repeated | SSO login options |






<a name="api-v1-authentication-GetRequest"></a>

### GetRequest
Get request to get the logged in user. The user ID etc. is taken from the
token.






<a name="api-v1-authentication-GetResponse"></a>

### GetResponse
Response with user information to the get request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_info | [model.UserInfo](#model-UserInfo) |  | Information about the user. |
| user_id | [string](#string) |  | ID of the user |






<a name="api-v1-authentication-LoginRequest"></a>

### LoginRequest
Login request with either user identifier & email or client credentials.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_password | [UserPassword](#api-v1-authentication-UserPassword) |  | User identifier & password. |
| client_credentials | [ClientCredentials](#api-v1-authentication-ClientCredentials) |  | Client credentials from service account. |






<a name="api-v1-authentication-LoginResponse"></a>

### LoginResponse
Login response with tokens and user information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| token | [Token](#api-v1-authentication-Token) |  | The access token and refresh token. |
| user_id | [string](#string) |  | ID of the user. |
| user_info | [model.UserInfo](#model-UserInfo) |  | User information. |






<a name="api-v1-authentication-LogoutRequest"></a>

### LogoutRequest
Empty logout request. The user ID etc. is taken from the token.






<a name="api-v1-authentication-LogoutResponse"></a>

### LogoutResponse
Empty response to the logout request.






<a name="api-v1-authentication-RefreshTokenRequest"></a>

### RefreshTokenRequest
Request to refresh the access and refresh token of the logged in user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refresh_token | [string](#string) |  | The access token of the user Refresh token matching the access token. |






<a name="api-v1-authentication-RefreshTokenResponse"></a>

### RefreshTokenResponse
Response with new access and refresh token.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| token | [Token](#api-v1-authentication-Token) |  | New refresh and access tokens |






<a name="api-v1-authentication-RegisterRequest"></a>

### RegisterRequest
Register request for users to self-register. This is only possible with the
register bool set in users settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_password | [UserPassword](#api-v1-authentication-UserPassword) |  | User identifier & password for the new user. |






<a name="api-v1-authentication-RegisterResponse"></a>

### RegisterResponse
Register response with tokens and user information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| token | [Token](#api-v1-authentication-Token) |  | Access and refresh token for the new logged in user. |
| user_id | [string](#string) |  | User ID of the new user. |
| user_info | [model.UserInfo](#model-UserInfo) |  | Information about the new user. |






<a name="api-v1-authentication-ResetPasswordRequest"></a>

### ResetPasswordRequest
Request to reset the password of a user with a verification code sent to the
email.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [string](#string) |  | The 6 digit verification code |
| new_password | [string](#string) |  | The new password |
| identifier | [model.UserIdentifier](#model-UserIdentifier) |  | Identifier of the user |
| token | [string](#string) |  | JWT token to reset the password. |






<a name="api-v1-authentication-ResetPasswordResponse"></a>

### ResetPasswordResponse
Empty response to the reset password request






<a name="api-v1-authentication-SSOID"></a>

### SSOID
Represents an SSO provided ID of a user


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [SSOType](#api-v1-authentication-SSOType) |  | What type of SSO this ID is from |
| provider_id | [string](#string) |  | The ID of the SSO provider |
| sso_id | [string](#string) |  | The ID provided by SSO |
| user_id | [string](#string) |  | The internal user ID |






<a name="api-v1-authentication-SSOOption"></a>

### SSOOption
A login option for using SSO. This might be merged into
GetAuthConfigResponse.login_types, but is introduced as a separate field, to
maintain backwards compatibility.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [SSOType](#api-v1-authentication-SSOType) |  | Type of SSO. Currently only OIDC is supported. |
| provider_id | [string](#string) |  | ID of the SSO provider as given in the platform configuration. |
| name | [string](#string) |  | Name of SSO provider. This is an optional human readable version of the provider ID. |
| icon | [string](#string) |  | URL of the underlying issuer. This can be used in the frontend for showing specific items for certain known issuers. |






<a name="api-v1-authentication-SendPasswordResetRequest"></a>

### SendPasswordResetRequest
Request to send a reset password email to the user. This is only possible if
an email provider is configured, and the user has an email.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identifier | [model.UserIdentifier](#model-UserIdentifier) |  | User identifier of the user. |






<a name="api-v1-authentication-SendPasswordResetResponse"></a>

### SendPasswordResetResponse
Empty response to the send password reset request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| token | [string](#string) |  | JWT token to reset the password. |






<a name="api-v1-authentication-SendVerificationEmailRequest"></a>

### SendVerificationEmailRequest
Request to send an email containing the code for the email verification flow.
This is an upsert, and will invalidate the current verification-code if it
exists. Only possible if an email-provider is configured, and the user has en
email.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identifier | [model.UserIdentifier](#model-UserIdentifier) |  | User identifier of the user. |






<a name="api-v1-authentication-SendVerificationEmailResponse"></a>

### SendVerificationEmailResponse
Empty response for sending a verification email


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| token | [string](#string) |  | JWT token to verify the email. |






<a name="api-v1-authentication-VerifyEmailRequest"></a>

### VerifyEmailRequest
Request to verify the email of a user with a verification code sent to the
email.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [string](#string) |  | The verification code. |
| email | [string](#string) |  | The email of the user. |
| token | [string](#string) |  | JWT token to verify the email. |






<a name="api-v1-authentication-VerifyEmailResponse"></a>

### VerifyEmailResponse
Empty response to the Verify Email Request.






<a name="api-v1-authentication-VerifyPhoneNumberRequest"></a>

### VerifyPhoneNumberRequest
Request to verify the phone number of a user with a verification code sent to
the phone number.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [string](#string) |  |  |
| phone_number | [string](#string) |  |  |






<a name="api-v1-authentication-VerifyPhoneNumberResponse"></a>

### VerifyPhoneNumberResponse
Empty response to the Verify Phone Number Request.








<a name="api-v1-authentication-SSOType"></a>

### SSOType
The type of SSO. Currently only OIDC is supported.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SSO_TYPE_UNSPECIFIED | 0 |  |
| SSO_TYPE_OIDC | 1 |  |








<a name="model_environment-proto"></a>

## model/environment.proto



<a name="model-EnvironmentFilter"></a>

### EnvironmentFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| all | [EnvironmentFilter.All](#model-EnvironmentFilter-All) |  |  |
| selected | [EnvironmentFilter.Selected](#model-EnvironmentFilter-Selected) |  |  |






<a name="model-EnvironmentFilter-All"></a>

### EnvironmentFilter.All



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| include_ephemeral | [bool](#bool) |  |  |






<a name="model-EnvironmentFilter-Selected"></a>

### EnvironmentFilter.Selected



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_ids | [string](#string) | repeated |  |













<a name="model_git-proto"></a>

## model/git.proto



<a name="model-BitBucketInfo"></a>

### BitBucketInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| team | [string](#string) |  |  |
| project | [string](#string) |  |  |
| repository | [string](#string) |  |  |






<a name="model-Commit"></a>

### Commit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| provider | [GitProvider](#model-GitProvider) |  |  |
| github | [GitHubInfo](#model-GitHubInfo) |  |  |
| gitlab | [GitLabInfo](#model-GitLabInfo) |  |  |
| bitbucket | [BitBucketInfo](#model-BitBucketInfo) |  |  |
| commit_id | [string](#string) |  |  |
| commit_url | [string](#string) |  |  |
| repository_url | [string](#string) |  |  |






<a name="model-GitChange"></a>

### GitChange



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#model-Commit) |  |  |
| paths | [string](#string) | repeated |  |






<a name="model-GitHubInfo"></a>

### GitHubInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| owner | [string](#string) |  |  |
| repository | [string](#string) |  |  |






<a name="model-GitLabInfo"></a>

### GitLabInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| groups | [string](#string) | repeated |  |
| project | [string](#string) |  |  |






<a name="model-GitStatus"></a>

### GitStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| last_processed_commit_id | [string](#string) |  |  |
| last_processed_commit_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| last_successful_commit_id | [string](#string) |  |  |
| last_successful_commit_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| error | [string](#string) |  |  |






<a name="model-GitStore"></a>

### GitStore



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| disabled | [bool](#bool) |  |  |
| repository | [string](#string) |  |  |
| branch | [string](#string) |  |  |
| capsule_path | [string](#string) |  |  |
| capsule_set_path | [string](#string) |  |  |
| commit_template | [string](#string) |  |  |
| environments | [EnvironmentFilter](#model-EnvironmentFilter) |  |  |
| pr_title_template | [string](#string) |  |  |
| pr_body_template | [string](#string) |  |  |
| require_pull_request | [bool](#bool) |  |  |






<a name="model-RepoBranch"></a>

### RepoBranch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repository | [string](#string) |  |  |
| branch | [string](#string) |  |  |








<a name="model-GitProvider"></a>

### GitProvider


| Name | Number | Description |
| ---- | ------ | ----------- |
| GIT_PROVIDER_UNSPECIFIED | 0 |  |
| GIT_PROVIDER_GITHUB | 1 |  |
| GIT_PROVIDER_GITLAB | 2 |  |
| GIT_PROVIDER_BITBUCKET | 3 |  |








<a name="api_v1_capsule_capsule-proto"></a>

## api/v1/capsule/capsule.proto



<a name="api-v1-capsule-Capsule"></a>

### Capsule
Environment wide capsule abstraction.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | ID of the capsule. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Last time the capsule was updated. |
| updated_by | [model.Author](#model-Author) |  | Author of the last update. |
| git_store | [model.GitStore](#model-GitStore) |  |  |






<a name="api-v1-capsule-Update"></a>

### Update



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| set_git_store | [model.GitStore](#model-GitStore) |  | Set the git store. |













<a name="api_v1_capsule_job-proto"></a>

## api/v1/capsule/job.proto



<a name="api-v1-capsule-CronJob"></a>

### CronJob
Specification for a cron job.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job_name | [string](#string) |  | Name of the job. |
| schedule | [string](#string) |  | Cron schedule. |
| max_retries | [int32](#int32) |  | Maximum number of retries. |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  | Maximum duration of the job. |
| url | [JobURL](#api-v1-capsule-JobURL) |  | URL job. |
| command | [JobCommand](#api-v1-capsule-JobCommand) |  | Command job. |






<a name="api-v1-capsule-JobCommand"></a>

### JobCommand
Run a job by running a command in an instance of a capsule


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| command | [string](#string) |  | Command to run. |
| args | [string](#string) | repeated | Arguments to pass to the command. |






<a name="api-v1-capsule-JobExecution"></a>

### JobExecution
An execution of a cron job.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job_name | [string](#string) |  | Name of the job. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the job started running. |
| finished_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the job finished. |
| state | [JobState](#api-v1-capsule-JobState) |  | The state of the job. |
| retries | [int32](#int32) |  | Number of retries. |
| rollout_id | [uint64](#uint64) |  | ID of the rollout. |
| capsule_id | [string](#string) |  | ID of the capsule. |
| project_id | [string](#string) |  | ID of the project. |
| execution_id | [string](#string) |  | ID of the execution. |
| environment_id | [string](#string) |  | ID of the environment. |






<a name="api-v1-capsule-JobURL"></a>

### JobURL
Run a job by making a HTTP request to a URL.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint64](#uint64) |  | Port to make the request to. |
| path | [string](#string) |  | Path to make the request to. |
| query_parameters | [JobURL.QueryParametersEntry](#api-v1-capsule-JobURL-QueryParametersEntry) | repeated | Query parameters to add to the request. |






<a name="api-v1-capsule-JobURL-QueryParametersEntry"></a>

### JobURL.QueryParametersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |








<a name="api-v1-capsule-JobState"></a>

### JobState
Different states a job execution can be in

| Name | Number | Description |
| ---- | ------ | ----------- |
| JOB_STATE_UNSPECIFIED | 0 | Default value. |
| JOB_STATE_ONGOING | 1 | The job is running. |
| JOB_STATE_COMPLETED | 2 | The job completed successfully. |
| JOB_STATE_FAILED | 3 | The job failed. |
| JOB_STATE_TERMINATED | 4 | The job was terminated. |








<a name="model_metrics-proto"></a>

## model/metrics.proto



<a name="model-ContainerMetrics"></a>

### ContainerMetrics
Metrics for a container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp of the metrics. |
| memory_bytes | [uint64](#uint64) |  | Memory usage in bytes. |
| cpu_ms | [uint64](#uint64) |  | CPU usage in milliseconds. |
| storage_bytes | [uint64](#uint64) |  | Storage usage in bytes. |






<a name="model-InstanceMetrics"></a>

### InstanceMetrics
Metrics for an instance


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule of the instance. |
| instance_id | [string](#string) |  | Instance ID. |
| main_container | [ContainerMetrics](#model-ContainerMetrics) |  | Main container metrics. |
| proxy_container | [ContainerMetrics](#model-ContainerMetrics) |  | Proxy container metrics. |






<a name="model-Metric"></a>

### Metric
Custom metrics


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the metric. |
| latest_value | [double](#double) |  | Latest value of the metric. |
| latest_timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp of the latest value. |






<a name="model-ObjectReference"></a>

### ObjectReference
A reference to a kubernetes object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  | Type of object. |
| name | [string](#string) |  | Name of the object. |
| api_version | [string](#string) |  | Api version of the object. |













<a name="k8s-io_api_autoscaling_v2_generated-proto"></a>

## k8s.io/api/autoscaling/v2/generated.proto



<a name="k8s-io-api-autoscaling-v2-CrossVersionObjectReference"></a>

### CrossVersionObjectReference



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  |  |
| name | [string](#string) |  |  |
| apiVersion | [string](#string) |  |  |













<a name="platform_v1_generated-proto"></a>

## platform/v1/generated.proto



<a name="platform-v1-CPUTarget"></a>

### CPUTarget



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| utilization | [uint32](#uint32) |  |  |






<a name="platform-v1-Capsule"></a>

### Capsule



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  |  |
| apiVersion | [string](#string) |  |  |
| name | [string](#string) |  |  |
| project | [string](#string) |  |  |
| environment | [string](#string) |  |  |
| spec | [CapsuleSpec](#platform-v1-CapsuleSpec) |  |  |






<a name="platform-v1-CapsuleInterface"></a>

### CapsuleInterface



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| port | [int32](#int32) |  |  |
| liveness | [InterfaceLivenessProbe](#platform-v1-InterfaceLivenessProbe) |  |  |
| readiness | [InterfaceReadinessProbe](#platform-v1-InterfaceReadinessProbe) |  |  |
| routes | [HostRoute](#platform-v1-HostRoute) | repeated |  |






<a name="platform-v1-CapsuleSet"></a>

### CapsuleSet



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  |  |
| apiVersion | [string](#string) |  |  |
| name | [string](#string) |  |  |
| project | [string](#string) |  |  |
| spec | [CapsuleSpec](#platform-v1-CapsuleSpec) |  |  |
| environments | [CapsuleSet.EnvironmentsEntry](#platform-v1-CapsuleSet-EnvironmentsEntry) | repeated |  |
| environmentRefs | [string](#string) | repeated |  |






<a name="platform-v1-CapsuleSet-EnvironmentsEntry"></a>

### CapsuleSet.EnvironmentsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [CapsuleSpec](#platform-v1-CapsuleSpec) |  |  |






<a name="platform-v1-CapsuleSpec"></a>

### CapsuleSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| annotations | [CapsuleSpec.AnnotationsEntry](#platform-v1-CapsuleSpec-AnnotationsEntry) | repeated |  |
| image | [string](#string) |  |  |
| command | [string](#string) |  |  |
| args | [string](#string) | repeated |  |
| interfaces | [CapsuleInterface](#platform-v1-CapsuleInterface) | repeated |  |
| files | [File](#platform-v1-File) | repeated |  |
| env | [EnvironmentVariables](#platform-v1-EnvironmentVariables) |  |  |
| scale | [Scale](#platform-v1-Scale) |  |  |
| cronJobs | [CronJob](#platform-v1-CronJob) | repeated |  |
| autoAddRigServiceAccounts | [bool](#bool) |  |  |
| extensions | [CapsuleSpec.ExtensionsEntry](#platform-v1-CapsuleSpec-ExtensionsEntry) | repeated |  |






<a name="platform-v1-CapsuleSpec-AnnotationsEntry"></a>

### CapsuleSpec.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="platform-v1-CapsuleSpec-ExtensionsEntry"></a>

### CapsuleSpec.ExtensionsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.Struct](#google-protobuf-Struct) |  |  |






<a name="platform-v1-CronJob"></a>

### CronJob



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| schedule | [string](#string) |  |  |
| url | [URL](#platform-v1-URL) |  |  |
| command | [JobCommand](#platform-v1-JobCommand) |  |  |
| maxRetries | [uint64](#uint64) |  |  |
| timeoutSeconds | [uint64](#uint64) |  |  |






<a name="platform-v1-CustomMetric"></a>

### CustomMetric



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instanceMetric | [InstanceMetric](#platform-v1-InstanceMetric) |  |  |
| objectMetric | [ObjectMetric](#platform-v1-ObjectMetric) |  |  |






<a name="platform-v1-Environment"></a>

### Environment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  |  |
| apiVersion | [string](#string) |  |  |
| name | [string](#string) |  |  |
| namespaceTemplate | [string](#string) |  |  |
| operatorVersion | [string](#string) |  |  |
| cluster | [string](#string) |  |  |
| spec | [ProjEnvCapsuleBase](#platform-v1-ProjEnvCapsuleBase) |  |  |
| ephemeral | [bool](#bool) |  |  |
| activeProjects | [string](#string) | repeated |  |
| global | [bool](#bool) |  |  |






<a name="platform-v1-EnvironmentSource"></a>

### EnvironmentSource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| kind | [string](#string) |  |  |






<a name="platform-v1-EnvironmentVariables"></a>

### EnvironmentVariables



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| raw | [EnvironmentVariables.RawEntry](#platform-v1-EnvironmentVariables-RawEntry) | repeated |  |
| sources | [EnvironmentSource](#platform-v1-EnvironmentSource) | repeated |  |






<a name="platform-v1-EnvironmentVariables-RawEntry"></a>

### EnvironmentVariables.RawEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="platform-v1-File"></a>

### File



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |
| asSecret | [bool](#bool) |  |  |
| bytes | [bytes](#bytes) |  |  |
| string | [string](#string) |  |  |
| ref | [FileReference](#platform-v1-FileReference) |  |  |






<a name="platform-v1-FileReference"></a>

### FileReference



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  |  |
| name | [string](#string) |  |  |
| key | [string](#string) |  |  |






<a name="platform-v1-HTTPPathRoute"></a>

### HTTPPathRoute



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |
| match | [string](#string) |  |  |






<a name="platform-v1-HorizontalScale"></a>

### HorizontalScale



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| min | [uint32](#uint32) |  |  |
| max | [uint32](#uint32) |  |  |
| instances | [Instances](#platform-v1-Instances) |  |  |
| cpuTarget | [CPUTarget](#platform-v1-CPUTarget) |  |  |
| customMetrics | [CustomMetric](#platform-v1-CustomMetric) | repeated |  |






<a name="platform-v1-HostCapsule"></a>

### HostCapsule



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  |  |
| apiVersion | [string](#string) |  |  |
| name | [string](#string) |  |  |
| project | [string](#string) |  |  |
| environment | [string](#string) |  |  |
| network | [HostNetwork](#platform-v1-HostNetwork) |  |  |






<a name="platform-v1-HostNetwork"></a>

### HostNetwork



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostInterfaces | [ProxyInterface](#platform-v1-ProxyInterface) | repeated |  |
| capsuleInterfaces | [ProxyInterface](#platform-v1-ProxyInterface) | repeated |  |
| tunnelPort | [uint32](#uint32) |  |  |






<a name="platform-v1-HostRoute"></a>

### HostRoute



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| host | [string](#string) |  |  |
| paths | [HTTPPathRoute](#platform-v1-HTTPPathRoute) | repeated |  |
| annotations | [HostRoute.AnnotationsEntry](#platform-v1-HostRoute-AnnotationsEntry) | repeated |  |






<a name="platform-v1-HostRoute-AnnotationsEntry"></a>

### HostRoute.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="platform-v1-InstanceMetric"></a>

### InstanceMetric



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metricName | [string](#string) |  |  |
| matchLabels | [InstanceMetric.MatchLabelsEntry](#platform-v1-InstanceMetric-MatchLabelsEntry) | repeated |  |
| averageValue | [string](#string) |  |  |






<a name="platform-v1-InstanceMetric-MatchLabelsEntry"></a>

### InstanceMetric.MatchLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="platform-v1-Instances"></a>

### Instances



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| min | [uint32](#uint32) |  |  |
| max | [uint32](#uint32) |  |  |






<a name="platform-v1-InterfaceGRPCProbe"></a>

### InterfaceGRPCProbe



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service | [string](#string) |  |  |
| enabled | [bool](#bool) |  |  |






<a name="platform-v1-InterfaceLivenessProbe"></a>

### InterfaceLivenessProbe



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |
| tcp | [bool](#bool) |  |  |
| grpc | [InterfaceGRPCProbe](#platform-v1-InterfaceGRPCProbe) |  |  |
| startupDelay | [uint32](#uint32) |  |  |






<a name="platform-v1-InterfaceOptions"></a>

### InterfaceOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tcp | [bool](#bool) |  |  |
| allowOrigin | [string](#string) |  |  |
| changeOrigin | [bool](#bool) |  |  |
| headers | [InterfaceOptions.HeadersEntry](#platform-v1-InterfaceOptions-HeadersEntry) | repeated |  |






<a name="platform-v1-InterfaceOptions-HeadersEntry"></a>

### InterfaceOptions.HeadersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="platform-v1-InterfaceReadinessProbe"></a>

### InterfaceReadinessProbe



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |
| tcp | [bool](#bool) |  |  |
| grpc | [InterfaceGRPCProbe](#platform-v1-InterfaceGRPCProbe) |  |  |






<a name="platform-v1-JobCommand"></a>

### JobCommand



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| command | [string](#string) |  |  |
| args | [string](#string) | repeated |  |






<a name="platform-v1-ObjectMetric"></a>

### ObjectMetric



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metricName | [string](#string) |  |  |
| matchLabels | [ObjectMetric.MatchLabelsEntry](#platform-v1-ObjectMetric-MatchLabelsEntry) | repeated |  |
| averageValue | [string](#string) |  |  |
| value | [string](#string) |  |  |
| objectReference | [k8s.io.api.autoscaling.v2.CrossVersionObjectReference](#k8s-io-api-autoscaling-v2-CrossVersionObjectReference) |  |  |






<a name="platform-v1-ObjectMetric-MatchLabelsEntry"></a>

### ObjectMetric.MatchLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="platform-v1-ProjEnvCapsuleBase"></a>

### ProjEnvCapsuleBase



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| files | [File](#platform-v1-File) | repeated |  |
| env | [EnvironmentVariables](#platform-v1-EnvironmentVariables) |  |  |






<a name="platform-v1-Project"></a>

### Project



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  |  |
| apiVersion | [string](#string) |  |  |
| name | [string](#string) |  |  |
| spec | [ProjEnvCapsuleBase](#platform-v1-ProjEnvCapsuleBase) |  |  |






<a name="platform-v1-ProxyInterface"></a>

### ProxyInterface



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  |  |
| target | [string](#string) |  |  |
| options | [InterfaceOptions](#platform-v1-InterfaceOptions) |  |  |






<a name="platform-v1-ResourceLimits"></a>

### ResourceLimits



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request | [string](#string) |  |  |
| limit | [string](#string) |  |  |






<a name="platform-v1-ResourceRequest"></a>

### ResourceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request | [string](#string) |  |  |






<a name="platform-v1-Scale"></a>

### Scale



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| horizontal | [HorizontalScale](#platform-v1-HorizontalScale) |  |  |
| vertical | [VerticalScale](#platform-v1-VerticalScale) |  |  |






<a name="platform-v1-URL"></a>

### URL



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  |  |
| path | [string](#string) |  |  |
| queryParameters | [URL.QueryParametersEntry](#platform-v1-URL-QueryParametersEntry) | repeated |  |






<a name="platform-v1-URL-QueryParametersEntry"></a>

### URL.QueryParametersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="platform-v1-VerticalScale"></a>

### VerticalScale



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cpu | [ResourceLimits](#platform-v1-ResourceLimits) |  |  |
| memory | [ResourceLimits](#platform-v1-ResourceLimits) |  |  |
| gpu | [ResourceRequest](#platform-v1-ResourceRequest) |  |  |













<a name="api_v1_capsule_change-proto"></a>

## api/v1/capsule/change.proto



<a name="api-v1-capsule-CPUTarget"></a>

### CPUTarget
Autoscaling based on CPU target.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| average_utilization_percentage | [uint32](#uint32) |  | Average CPU utilization target. |






<a name="api-v1-capsule-Change"></a>

### Change
Change to a capsule that ultimately results in a new rollout.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| replicas | [uint32](#uint32) |  | Number of replicas changed. |
| image_id | [string](#string) |  | New image change. |
| network | [Network](#api-v1-capsule-Network) |  | Network interfaces change. |
| container_settings | [ContainerSettings](#api-v1-capsule-ContainerSettings) |  | Container settings of the instances. |
| auto_add_rig_service_accounts | [bool](#bool) |  | Automatically add a rig-service account. |
| set_config_file | [Change.ConfigFile](#api-v1-capsule-Change-ConfigFile) |  | Set a config file - either update or add. |
| set_config_file_ref | [Change.ConfigFileRef](#api-v1-capsule-Change-ConfigFileRef) |  | Set a config file ref - either update or add. |
| remove_config_file | [string](#string) |  | Path of a config file to remove. |
| horizontal_scale | [HorizontalScale](#api-v1-capsule-HorizontalScale) |  | Horizontal scaling settings. |
| rollback | [Change.Rollback](#api-v1-capsule-Change-Rollback) |  | Rollback to a previous rollout. |
| set_environment_variable | [Change.KeyValue](#api-v1-capsule-Change-KeyValue) |  | Update or add an environment variable. |
| remove_environment_variable | [string](#string) |  | Name of an environment variable to remove. |
| set_environment_source | [EnvironmentSource](#api-v1-capsule-EnvironmentSource) |  | Set or update an environment source. |
| remove_environment_source | [EnvironmentSource](#api-v1-capsule-EnvironmentSource) |  | Remove an environment source. |
| command_arguments | [Change.CommandArguments](#api-v1-capsule-Change-CommandArguments) |  | Entrypoint for capsule instances. |
| add_cron_job | [CronJob](#api-v1-capsule-CronJob) |  | Add a cron job. |
| remove_cron_job | [Change.RemoveCronJob](#api-v1-capsule-Change-RemoveCronJob) |  | Remove a cron job. |
| set_interface | [Interface](#api-v1-capsule-Interface) |  | Add or update a network interface. |
| remove_interface | [string](#string) |  | Remove a network interface. |
| set_annotations | [Change.Annotations](#api-v1-capsule-Change-Annotations) |  | Set capsule annotations. |
| set_annotation | [Change.KeyValue](#api-v1-capsule-Change-KeyValue) |  | Update or add a single capsule annotation. |
| remove_annotation | [string](#string) |  | Name of a single capsule annotation to remove. |
| add_image | [Change.AddImage](#api-v1-capsule-Change-AddImage) |  | Image to deploy, adding it to images if not already present. |
| spec | [platform.v1.CapsuleSpec](#platform-v1-CapsuleSpec) |  | Complete capsule-spec to replace the current. |






<a name="api-v1-capsule-Change-AddImage"></a>

### Change.AddImage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| image | [string](#string) |  |  |






<a name="api-v1-capsule-Change-Annotations"></a>

### Change.Annotations



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| annotations | [Change.Annotations.AnnotationsEntry](#api-v1-capsule-Change-Annotations-AnnotationsEntry) | repeated |  |






<a name="api-v1-capsule-Change-Annotations-AnnotationsEntry"></a>

### Change.Annotations.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="api-v1-capsule-Change-CommandArguments"></a>

### Change.CommandArguments
Entrypoint for the capsule instances.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| command | [string](#string) |  | Command to run. |
| args | [string](#string) | repeated | arguments to the command. |






<a name="api-v1-capsule-Change-ConfigFile"></a>

### Change.ConfigFile
Config file change.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  | Path of the file in the instance. |
| content | [bytes](#bytes) |  | Content of the config file. |
| is_secret | [bool](#bool) |  | True if the content is secret. |






<a name="api-v1-capsule-Change-ConfigFileRef"></a>

### Change.ConfigFileRef
Config file ref change.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  | Path of the file in the instance. |
| kind | [string](#string) |  | Kind of the object to inject as config file. Either ConfigMap or Secret. |
| name | [string](#string) |  | Name of the object to inject as a config file. |
| key | [string](#string) |  | Key of the data within the object contents. |






<a name="api-v1-capsule-Change-CronJobs"></a>

### Change.CronJobs
Jobs change


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| jobs | [CronJob](#api-v1-capsule-CronJob) | repeated | The jobs. |






<a name="api-v1-capsule-Change-KeyValue"></a>

### Change.KeyValue
Key-value change.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the property. |
| value | [string](#string) |  | The value of the property. |






<a name="api-v1-capsule-Change-RemoveCronJob"></a>

### Change.RemoveCronJob
Remove cron job change.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job_name | [string](#string) |  | Name of the job to remove |






<a name="api-v1-capsule-Change-Rollback"></a>

### Change.Rollback
Rollback change.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rollback_id | [uint64](#uint64) |  | Rollout to roll back to. |






<a name="api-v1-capsule-ContainerSettings"></a>

### ContainerSettings
Settings for the instance container


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_variables | [ContainerSettings.EnvironmentVariablesEntry](#api-v1-capsule-ContainerSettings-EnvironmentVariablesEntry) | repeated | Environment variables. |
| command | [string](#string) |  | Entrypoint for the container. |
| args | [string](#string) | repeated | Arguments to the container. |
| resources | [Resources](#api-v1-capsule-Resources) |  | Resource requests and limits. |
| environment_sources | [EnvironmentSource](#api-v1-capsule-EnvironmentSource) | repeated | Environment sources. |






<a name="api-v1-capsule-ContainerSettings-EnvironmentVariablesEntry"></a>

### ContainerSettings.EnvironmentVariablesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="api-v1-capsule-CustomMetric"></a>

### CustomMetric
Autoscaling based on custom metrics.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance | [InstanceMetric](#api-v1-capsule-InstanceMetric) |  | If scaling based on metrics across all intstances / pods. |
| object | [ObjectMetric](#api-v1-capsule-ObjectMetric) |  | If scaling based on metrics for a specific kubernetes resource object. |






<a name="api-v1-capsule-EnvironmentSource"></a>

### EnvironmentSource
Source of environment variables


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the source |
| kind | [EnvironmentSource.Kind](#api-v1-capsule-EnvironmentSource-Kind) |  | Type of the source |






<a name="api-v1-capsule-GpuLimits"></a>

### GpuLimits
GPU resource limits


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  | gpu resource type - fx. nvidia.com/gpu |
| count | [uint32](#uint32) |  | number of gpus |






<a name="api-v1-capsule-HTTPPathRoute"></a>

### HTTPPathRoute



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |
| match | [PathMatchType](#api-v1-capsule-PathMatchType) |  |  |






<a name="api-v1-capsule-HorizontalScale"></a>

### HorizontalScale
Horizontal scaling settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_replicas | [uint32](#uint32) |  | Maximum number of replicas. |
| min_replicas | [uint32](#uint32) |  | Minimum number of replicas. |
| cpu_target | [CPUTarget](#api-v1-capsule-CPUTarget) |  | CPU target for autoscaling. |
| custom_metrics | [CustomMetric](#api-v1-capsule-CustomMetric) | repeated | If scaling based on custom metrics. |






<a name="api-v1-capsule-HostRoute"></a>

### HostRoute



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | [string](#string) |  |  |
| options | [RouteOptions](#api-v1-capsule-RouteOptions) |  |  |
| paths | [HTTPPathRoute](#api-v1-capsule-HTTPPathRoute) | repeated |  |
| id | [string](#string) |  |  |






<a name="api-v1-capsule-InstanceMetric"></a>

### InstanceMetric
Metric emitted by instances / pods.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metric_name | [string](#string) |  | Name of the metric |
| match_labels | [InstanceMetric.MatchLabelsEntry](#api-v1-capsule-InstanceMetric-MatchLabelsEntry) | repeated | Labels of the instances to match. |
| average_value | [string](#string) |  | Average value target. |






<a name="api-v1-capsule-InstanceMetric-MatchLabelsEntry"></a>

### InstanceMetric.MatchLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="api-v1-capsule-Interface"></a>

### Interface
A single network interface.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  | Port of the interface. |
| name | [string](#string) |  | Name of the interface. |
| public | [PublicInterface](#api-v1-capsule-PublicInterface) |  | If public interface is enabled. Contains ingress or load balancer settings. |
| liveness | [InterfaceProbe](#api-v1-capsule-InterfaceProbe) |  | Liveness probe. |
| readiness | [InterfaceProbe](#api-v1-capsule-InterfaceProbe) |  | Readiness probe. |
| routes | [HostRoute](#api-v1-capsule-HostRoute) | repeated | Routes for the network interface. |






<a name="api-v1-capsule-InterfaceProbe"></a>

### InterfaceProbe
Probe for liveness or readiness.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| http | [InterfaceProbe.HTTP](#api-v1-capsule-InterfaceProbe-HTTP) |  |  |
| tcp | [InterfaceProbe.TCP](#api-v1-capsule-InterfaceProbe-TCP) |  |  |
| grpc | [InterfaceProbe.GRPC](#api-v1-capsule-InterfaceProbe-GRPC) |  |  |






<a name="api-v1-capsule-InterfaceProbe-GRPC"></a>

### InterfaceProbe.GRPC
GRPC service for the probe.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service | [string](#string) |  |  |






<a name="api-v1-capsule-InterfaceProbe-HTTP"></a>

### InterfaceProbe.HTTP
HTTP path for the probe.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |






<a name="api-v1-capsule-InterfaceProbe-TCP"></a>

### InterfaceProbe.TCP
IF TCP probe.






<a name="api-v1-capsule-Network"></a>

### Network
A network configuration of network interfaces.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| interfaces | [Interface](#api-v1-capsule-Interface) | repeated | All the network interfaces. |






<a name="api-v1-capsule-ObjectMetric"></a>

### ObjectMetric
Metric emitted by kubernetes object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metric_name | [string](#string) |  | Name of the metric. |
| match_labels | [ObjectMetric.MatchLabelsEntry](#api-v1-capsule-ObjectMetric-MatchLabelsEntry) | repeated | Labels of the object to match. |
| average_value | [string](#string) |  | Average value target. |
| value | [string](#string) |  | Value target. |
| object_reference | [model.ObjectReference](#model-ObjectReference) |  | Reference to the object. |






<a name="api-v1-capsule-ObjectMetric-MatchLabelsEntry"></a>

### ObjectMetric.MatchLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="api-v1-capsule-PublicInterface"></a>

### PublicInterface
Public interface configuration.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  | True if the interface is public. |
| method | [RoutingMethod](#api-v1-capsule-RoutingMethod) |  | Routing method - Ingress or Load Balancer. |






<a name="api-v1-capsule-ResourceList"></a>

### ResourceList
CPU and Memory resource request or limits


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cpu_millis | [uint32](#uint32) |  | Unit: milli-CPUs |
| memory_bytes | [uint64](#uint64) |  | Unit: Bytes |






<a name="api-v1-capsule-Resources"></a>

### Resources
Container resources requests and limits


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| requests | [ResourceList](#api-v1-capsule-ResourceList) |  | CPU and memory requests. |
| limits | [ResourceList](#api-v1-capsule-ResourceList) |  | CPU and memory limits |
| gpu_limits | [GpuLimits](#api-v1-capsule-GpuLimits) |  | GPU Limits |






<a name="api-v1-capsule-RouteOptions"></a>

### RouteOptions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| annotations | [RouteOptions.AnnotationsEntry](#api-v1-capsule-RouteOptions-AnnotationsEntry) | repeated |  |






<a name="api-v1-capsule-RouteOptions-AnnotationsEntry"></a>

### RouteOptions.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="api-v1-capsule-RoutingMethod"></a>

### RoutingMethod
The routing method for the public interface.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| load_balancer | [RoutingMethod.LoadBalancer](#api-v1-capsule-RoutingMethod-LoadBalancer) |  |  |
| ingress | [RoutingMethod.Ingress](#api-v1-capsule-RoutingMethod-Ingress) |  |  |






<a name="api-v1-capsule-RoutingMethod-Ingress"></a>

### RoutingMethod.Ingress
Ingress routing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | [string](#string) |  | Hostname of the ingress. |
| tls | [bool](#bool) |  | True if TLS is enabled. |
| paths | [string](#string) | repeated | Paths of the ingress. |






<a name="api-v1-capsule-RoutingMethod-LoadBalancer"></a>

### RoutingMethod.LoadBalancer
Loadbalancer routing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  | public port. |
| node_port | [uint32](#uint32) |  | node port. |








<a name="api-v1-capsule-EnvironmentSource-Kind"></a>

### EnvironmentSource.Kind


| Name | Number | Description |
| ---- | ------ | ----------- |
| KIND_UNSPECIFIED | 0 | Unspecified. |
| KIND_CONFIG_MAP | 1 | Environment variables from a config map. |
| KIND_SECRET | 2 | Environment variables from a secret. |



<a name="api-v1-capsule-PathMatchType"></a>

### PathMatchType


| Name | Number | Description |
| ---- | ------ | ----------- |
| PATH_MATCH_TYPE_UNSPECIFIED | 0 |  |
| PATH_MATCH_TYPE_PATH_PREFIX | 1 |  |
| PATH_MATCH_TYPE_EXACT | 2 |  |
| PATH_MATCH_TYPE_REGULAR_EXPRESSION | 3 |  |








<a name="api_v1_capsule_event-proto"></a>

## api/v1/capsule/event.proto



<a name="api-v1-capsule-AbortEvent"></a>

### AbortEvent
An event that is associated with an abort.






<a name="api-v1-capsule-ErrorEvent"></a>

### ErrorEvent
An event that is associated with an error.






<a name="api-v1-capsule-Event"></a>

### Event
An event is a message from a rollout


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created_by | [model.Author](#model-Author) |  | Potential author associated with the event. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the event was created. |
| rollout_id | [uint64](#uint64) |  | The rollout that created the event. |
| message | [string](#string) |  | A message associated with the event. |
| event_data | [EventData](#api-v1-capsule-EventData) |  | The data associated with the event. |






<a name="api-v1-capsule-EventData"></a>

### EventData
The data associated with an event.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rollout | [RolloutEvent](#api-v1-capsule-RolloutEvent) |  | If event is a rollout. |
| error | [ErrorEvent](#api-v1-capsule-ErrorEvent) |  | if event is an error event. |
| abort | [AbortEvent](#api-v1-capsule-AbortEvent) |  | If event is an abort event. |






<a name="api-v1-capsule-RolloutEvent"></a>

### RolloutEvent
An event that is associated with a rollout.













<a name="api_v1_capsule_field-proto"></a>

## api/v1/capsule/field.proto



<a name="api-v1-capsule-FieldChange"></a>

### FieldChange



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| field_id | [string](#string) |  | The Field-ID associated with this change. This is formated as a json-path string with '?' placeholders. |
| field_path | [string](#string) |  | The unique Field-path identifying this change. This is formated as a json-path string. |
| old_value_yaml | [string](#string) |  | YAML encoding of the old value (if any). |
| new_value_yaml | [string](#string) |  | YAML encoding of the new value (if any). |
| operation | [FieldOperation](#api-v1-capsule-FieldOperation) |  | Operation is how this field-change is applied against the source to get to the target spec. |
| description | [string](#string) |  | Pretty-printed string description of the field change. |








<a name="api-v1-capsule-FieldOperation"></a>

### FieldOperation


| Name | Number | Description |
| ---- | ------ | ----------- |
| FIELD_OPERATION_UNSPECIFIED | 0 |  |
| FIELD_OPERATION_ADDED | 1 |  |
| FIELD_OPERATION_REMOVED | 2 |  |
| FIELD_OPERATION_MODIFIED | 3 |  |








<a name="api_v1_capsule_image-proto"></a>

## api/v1/capsule/image.proto



<a name="api-v1-capsule-GitReference"></a>

### GitReference
GitReference is an origin of a image.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repository_url | [string](#string) |  | The url of the git repository |
| commit_sha | [string](#string) |  | The commit sha of the git repository |
| commit_url | [string](#string) |  | The commit url of the git repository |






<a name="api-v1-capsule-Image"></a>

### Image
Image is an cross-environment abstraction of an container image along with
metadata for a capsule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| image_id | [string](#string) |  | unique identifier for the image |
| digest | [string](#string) |  | digest of the image |
| repository | [string](#string) |  | repository of the image |
| tag | [string](#string) |  | tag of the image |
| created_by | [model.Author](#model-Author) |  | user who created the image |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | time the image was created |
| origin | [Origin](#api-v1-capsule-Origin) |  | origin of the image |
| labels | [Image.LabelsEntry](#api-v1-capsule-Image-LabelsEntry) | repeated | labels of the image |






<a name="api-v1-capsule-Image-LabelsEntry"></a>

### Image.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="api-v1-capsule-Origin"></a>

### Origin
Where the image came from


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| git_reference | [GitReference](#api-v1-capsule-GitReference) |  | The image came from a git repository |













<a name="api_v1_capsule_instance-proto"></a>

## api/v1/capsule/instance.proto



<a name="api-v1-capsule-ContainerStateTerminated"></a>

### ContainerStateTerminated



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| exit_code | [int32](#int32) |  | Exit status from the last termination of the container |
| signal | [int32](#int32) |  | Signal from the last termination of the container +optional |
| reason | [string](#string) |  | (brief) reason from the last termination of the container +optional |
| message | [string](#string) |  | Message regarding the last termination of the container +optional |
| started_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time at which previous execution of the container started +optional |
| finished_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time at which the container last terminated +optional |
| container_id | [string](#string) |  | Container's ID in the format 'type://container_id' +optional |






<a name="api-v1-capsule-CrashLoopBackoff"></a>

### CrashLoopBackoff



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |
| data | [CrashLoopBackoffData](#api-v1-capsule-CrashLoopBackoffData) |  |  |






<a name="api-v1-capsule-CrashLoopBackoffData"></a>

### CrashLoopBackoffData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| termination | [ContainerStateTerminated](#api-v1-capsule-ContainerStateTerminated) |  |  |






<a name="api-v1-capsule-CurrentlyUnscheduleable"></a>

### CurrentlyUnscheduleable



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |






<a name="api-v1-capsule-DoneScheduling"></a>

### DoneScheduling



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |






<a name="api-v1-capsule-ImagePulling"></a>

### ImagePulling



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |
| stages | [ImagePullingStages](#api-v1-capsule-ImagePullingStages) |  |  |






<a name="api-v1-capsule-ImagePullingBackOff"></a>

### ImagePullingBackOff



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |






<a name="api-v1-capsule-ImagePullingDone"></a>

### ImagePullingDone



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |






<a name="api-v1-capsule-ImagePullingError"></a>

### ImagePullingError



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |






<a name="api-v1-capsule-ImagePullingPulling"></a>

### ImagePullingPulling



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |






<a name="api-v1-capsule-ImagePullingStages"></a>

### ImagePullingStages



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pulling | [ImagePullingPulling](#api-v1-capsule-ImagePullingPulling) |  |  |
| error | [ImagePullingError](#api-v1-capsule-ImagePullingError) |  |  |
| back_off | [ImagePullingBackOff](#api-v1-capsule-ImagePullingBackOff) |  |  |
| done | [ImagePullingDone](#api-v1-capsule-ImagePullingDone) |  |  |






<a name="api-v1-capsule-Instance"></a>

### Instance



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance_id | [string](#string) |  |  |
| image_id | [string](#string) |  |  |
| state | [State](#api-v1-capsule-State) |  |  |
| restart_count | [uint32](#uint32) |  |  |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| started_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| finished_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| message | [string](#string) |  |  |
| rollout_id | [uint64](#uint64) |  |  |






<a name="api-v1-capsule-InstanceReady"></a>

### InstanceReady



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |






<a name="api-v1-capsule-InstanceStatus"></a>

### InstanceStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |
| data | [InstanceStatusData](#api-v1-capsule-InstanceStatusData) |  |  |
| stages | [InstanceStatusStages](#api-v1-capsule-InstanceStatusStages) |  |  |






<a name="api-v1-capsule-InstanceStatusData"></a>

### InstanceStatusData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance_id | [string](#string) |  |  |
| rollout_id | [uint64](#uint64) |  |  |
| image_name | [string](#string) |  |  |
| node | [string](#string) |  |  |






<a name="api-v1-capsule-InstanceStatusPreparing"></a>

### InstanceStatusPreparing
======================= PREPARING =====================


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |
| stages | [InstanceStatusPreparingStages](#api-v1-capsule-InstanceStatusPreparingStages) |  |  |






<a name="api-v1-capsule-InstanceStatusPreparingStages"></a>

### InstanceStatusPreparingStages



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pulling | [ImagePulling](#api-v1-capsule-ImagePulling) |  |  |






<a name="api-v1-capsule-InstanceStatusRunning"></a>

### InstanceStatusRunning
======================== RUNNING ======================


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |
| stages | [InstanceStatusRunningStages](#api-v1-capsule-InstanceStatusRunningStages) |  |  |
| data | [InstanceStatusRunningData](#api-v1-capsule-InstanceStatusRunningData) |  |  |






<a name="api-v1-capsule-InstanceStatusRunningData"></a>

### InstanceStatusRunningData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| restarts | [uint32](#uint32) |  |  |






<a name="api-v1-capsule-InstanceStatusRunningStages"></a>

### InstanceStatusRunningStages



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| crash_loop_backoff | [CrashLoopBackoff](#api-v1-capsule-CrashLoopBackoff) |  |  |
| ready | [Ready](#api-v1-capsule-Ready) |  |  |
| running | [Running](#api-v1-capsule-Running) |  |  |






<a name="api-v1-capsule-InstanceStatusScheduling"></a>

### InstanceStatusScheduling
====================== SCHEDULING ====================


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |
| stages | [InstanceStatusSchedulingStages](#api-v1-capsule-InstanceStatusSchedulingStages) |  |  |






<a name="api-v1-capsule-InstanceStatusSchedulingStages"></a>

### InstanceStatusSchedulingStages



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| currently_unscheduleable | [CurrentlyUnscheduleable](#api-v1-capsule-CurrentlyUnscheduleable) |  |  |
| done | [DoneScheduling](#api-v1-capsule-DoneScheduling) |  |  |






<a name="api-v1-capsule-InstanceStatusStages"></a>

### InstanceStatusStages



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| schedule | [InstanceStatusScheduling](#api-v1-capsule-InstanceStatusScheduling) |  |  |
| preparing | [InstanceStatusPreparing](#api-v1-capsule-InstanceStatusPreparing) |  |  |
| running | [InstanceStatusRunning](#api-v1-capsule-InstanceStatusRunning) |  |  |






<a name="api-v1-capsule-NotReady"></a>

### NotReady



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |






<a name="api-v1-capsule-Ready"></a>

### Ready



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |
| stages | [ReadyStages](#api-v1-capsule-ReadyStages) |  |  |






<a name="api-v1-capsule-ReadyStages"></a>

### ReadyStages



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| not_ready | [NotReady](#api-v1-capsule-NotReady) |  |  |
| ready | [InstanceReady](#api-v1-capsule-InstanceReady) |  |  |






<a name="api-v1-capsule-Running"></a>

### Running



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamps | [StatusTimestamps](#api-v1-capsule-StatusTimestamps) |  |  |






<a name="api-v1-capsule-StatusTimestamps"></a>

### StatusTimestamps



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entered | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| updated | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| exited | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |








<a name="api-v1-capsule-State"></a>

### State


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATE_UNSPECIFIED | 0 |  |
| STATE_PENDING | 1 |  |
| STATE_RUNNING | 2 |  |
| STATE_SUCCEEDED | 3 |  |
| STATE_FAILED | 4 |  |








<a name="api_v1_capsule_instance_status-proto"></a>

## api/v1/capsule/instance/status.proto



<a name="api-v1-capsule-instance-ContainerInfo"></a>

### ContainerInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| type | [api.v1.pipeline.ContainerType](#api-v1-pipeline-ContainerType) |  |  |






<a name="api-v1-capsule-instance-ContainerTermination"></a>

### ContainerTermination
Information about the last container termination.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| exit_code | [int32](#int32) |  | Exit status from the last termination of the container |
| signal | [int32](#int32) |  | Signal from the last termination of the container +optional |
| reason | [string](#string) |  | (brief) reason from the last termination of the container +optional |
| message | [string](#string) |  | Message regarding the last termination of the container +optional |
| started_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time at which previous execution of the container started +optional |
| finished_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time at which the container last terminated +optional |
| container_id | [string](#string) |  | Container's ID in the format 'type://container_id' +optional |






<a name="api-v1-capsule-instance-DeletedStage"></a>

### DeletedStage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StageInfo](#api-v1-capsule-instance-StageInfo) |  |  |
| steps | [DeletedStep](#api-v1-capsule-instance-DeletedStep) | repeated |  |






<a name="api-v1-capsule-instance-DeletedStep"></a>

### DeletedStep



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| generic | [GenericStep](#api-v1-capsule-instance-GenericStep) |  |  |






<a name="api-v1-capsule-instance-ExecutingStep"></a>

### ExecutingStep
An executing step of the running stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-instance-StepInfo) |  | Meta information about the step. |
| started_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time at which the step started. |
| finished_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time at which the step finished. |
| restarts | [uint32](#uint32) |  | Number of restarts of the container |
| last_container_termination | [ContainerTermination](#api-v1-capsule-instance-ContainerTermination) |  | Information about the last container termination. |






<a name="api-v1-capsule-instance-GenericStep"></a>

### GenericStep
A generic step.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-instance-StepInfo) |  |  |






<a name="api-v1-capsule-instance-ImagePullingStep"></a>

### ImagePullingStep
An image pulling step of the preparing stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-instance-StepInfo) |  | Meta information about the step. |
| state | [ImagePullingState](#api-v1-capsule-instance-ImagePullingState) |  | State of the step. |
| image | [string](#string) |  | Image that is being pulled. |






<a name="api-v1-capsule-instance-PlacementStep"></a>

### PlacementStep
Placement step.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-instance-StepInfo) |  | Meta information about the step. |
| node | [string](#string) |  | Node on which the instance should run. |






<a name="api-v1-capsule-instance-PreparingStage"></a>

### PreparingStage
The preparing stage


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StageInfo](#api-v1-capsule-instance-StageInfo) |  | Meta information about the stage. |
| steps | [PreparingStep](#api-v1-capsule-instance-PreparingStep) | repeated | Steps of the stage. |






<a name="api-v1-capsule-instance-PreparingStep"></a>

### PreparingStep
A step of the preparing stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| generic | [GenericStep](#api-v1-capsule-instance-GenericStep) |  | Generic step. |
| image_pulling | [ImagePullingStep](#api-v1-capsule-instance-ImagePullingStep) |  | Image pulling step. |
| init_executing | [ExecutingStep](#api-v1-capsule-instance-ExecutingStep) |  | Executing step for init containers |






<a name="api-v1-capsule-instance-ReadyStep"></a>

### ReadyStep
A ready step of the running stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StepInfo](#api-v1-capsule-instance-StepInfo) |  | Meta information about the step. |
| state | [ReadyState](#api-v1-capsule-instance-ReadyState) |  | State of the step. |
| failed_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time at which the step failed. |
| fail_count | [uint32](#uint32) |  | Number of times the step has failed. |






<a name="api-v1-capsule-instance-RunningStage"></a>

### RunningStage
The running stage of the instance


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StageInfo](#api-v1-capsule-instance-StageInfo) |  | Meta information about the stage. |
| steps | [RunningStep](#api-v1-capsule-instance-RunningStep) | repeated | Steps of the stage. |
| restarts | [uint32](#uint32) |  | Number of restarts of the instance. |
| last_container_termination | [ContainerTermination](#api-v1-capsule-instance-ContainerTermination) |  | Information about the last container termination. |






<a name="api-v1-capsule-instance-RunningStep"></a>

### RunningStep
A step of the running stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| generic | [GenericStep](#api-v1-capsule-instance-GenericStep) |  | Generic step. |
| ready | [ReadyStep](#api-v1-capsule-instance-ReadyStep) |  | Ready step. |
| executing | [ExecutingStep](#api-v1-capsule-instance-ExecutingStep) |  | Executing step. |






<a name="api-v1-capsule-instance-SchedulingStage"></a>

### SchedulingStage
The scheduling stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [StageInfo](#api-v1-capsule-instance-StageInfo) |  | Meta information about the stage. |
| steps | [SchedulingStep](#api-v1-capsule-instance-SchedulingStep) | repeated | Steps of the stage. |






<a name="api-v1-capsule-instance-SchedulingStep"></a>

### SchedulingStep
A step of the scheduling stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| generic | [GenericStep](#api-v1-capsule-instance-GenericStep) |  | Generic step. |
| placement | [PlacementStep](#api-v1-capsule-instance-PlacementStep) |  | Placement step - On what node should the instance run. |






<a name="api-v1-capsule-instance-StageInfo"></a>

### StageInfo
Meta information about a stage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the stage. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Last update time of the stage. |
| state | [StageState](#api-v1-capsule-instance-StageState) |  | State of the stage. |






<a name="api-v1-capsule-instance-Stages"></a>

### Stages
The different stages of the instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| schedule | [SchedulingStage](#api-v1-capsule-instance-SchedulingStage) |  | Scheduling stage. |
| preparing | [PreparingStage](#api-v1-capsule-instance-PreparingStage) |  | Preparing stage. |
| running | [RunningStage](#api-v1-capsule-instance-RunningStage) |  | Running stage. |
| deleted | [DeletedStage](#api-v1-capsule-instance-DeletedStage) |  | Deleted stage. |






<a name="api-v1-capsule-instance-Status"></a>

### Status
Status is a representation of the current state of an instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance_id | [string](#string) |  | Instance ID. |
| stages | [Stages](#api-v1-capsule-instance-Stages) |  | Stages of the instance. |
| rollout_id | [uint64](#uint64) |  | Rollout ID. |
| image | [string](#string) |  | Image of the instance. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Creation time of the instance. |






<a name="api-v1-capsule-instance-StepInfo"></a>

### StepInfo
Meta data about a step.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the step. |
| message | [string](#string) |  | Message of the step. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Last update time of the step. |
| state | [StepState](#api-v1-capsule-instance-StepState) |  | State of the step. |
| container | [ContainerInfo](#api-v1-capsule-instance-ContainerInfo) |  | Information about the container associated with the step. |








<a name="api-v1-capsule-instance-ImagePullingState"></a>

### ImagePullingState
Different states of an image pulling step.

| Name | Number | Description |
| ---- | ------ | ----------- |
| IMAGE_PULLING_STATE_UNSPECIFIED | 0 | Unspecified state. |
| IMAGE_PULLING_STATE_PULLING | 1 | Image is being pulled. |
| IMAGE_PULLING_STATE_ERROR | 2 | Image pulling has failed. |
| IMAGE_PULLING_STATE_BACK_OFF | 3 | Image pulling is in back off. |
| IMAGE_PULLING_STATE_DONE | 4 | Image pulling is done. |



<a name="api-v1-capsule-instance-PlacementState"></a>

### PlacementState
Different states of a placement step

| Name | Number | Description |
| ---- | ------ | ----------- |
| SCHEDULING_STATE_UNSPECIFIED | 0 | Unspecified state. |
| SCHEDULING_STATE_UNSCHEDULEABLE | 1 | If the instance is unschedulable. |
| SCHEDULING_STATE_DONE | 2 | If the instance is scheduled. |



<a name="api-v1-capsule-instance-ReadyState"></a>

### ReadyState
Different states of a ready step.

| Name | Number | Description |
| ---- | ------ | ----------- |
| READY_STATE_UNSPECIFIED | 0 | Unspecified state. |
| READY_STATE_CRASH_LOOP_BACKOFF | 1 | If the instance is in crash loop backoff. |
| READY_STATE_NOT_READY | 2 | If the instance is not ready. |
| READY_STATE_READY | 3 | If the instance is ready. |



<a name="api-v1-capsule-instance-StageState"></a>

### StageState
Different states a stage can be in.

| Name | Number | Description |
| ---- | ------ | ----------- |
| STAGE_STATE_UNSPECIFIED | 0 | Unspecified state. |
| STAGE_STATE_ONGOING | 1 | Stage is ongoing. |
| STAGE_STATE_FAILED | 2 | Stage has failed. |
| STAGE_STATE_DONE | 3 | Stage is done. |
| STAGE_STATE_RUNNING | 4 | Stage is running. |



<a name="api-v1-capsule-instance-StepState"></a>

### StepState
Different states a step can be in.

| Name | Number | Description |
| ---- | ------ | ----------- |
| STEP_STATE_UNSPECIFIED | 0 | Unspecified state. |
| STEP_STATE_ONGOING | 1 | Step is ongoing. |
| STEP_STATE_FAILED | 2 | Step has failed. |
| STEP_STATE_DONE | 3 | Step is done. |
| STEP_STATE_RUNNING | 4 | Step is running. |








<a name="api_v1_capsule_log-proto"></a>

## api/v1/capsule/log.proto



<a name="api-v1-capsule-Log"></a>

### Log
Log of an instance


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp of the log |
| message | [LogMessage](#api-v1-capsule-LogMessage) |  | Message of the log |
| instance_id | [string](#string) |  | Instance ID of the log |






<a name="api-v1-capsule-LogMessage"></a>

### LogMessage
The actual log message


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stdout | [bytes](#bytes) |  | If the log is stdout |
| stderr | [bytes](#bytes) |  | If the log is stderr |
| container_termination | [LogMessage.ContainerTermination](#api-v1-capsule-LogMessage-ContainerTermination) |  | Represents a termination event |






<a name="api-v1-capsule-LogMessage-ContainerTermination"></a>

### LogMessage.ContainerTermination














<a name="model_pipeline-proto"></a>

## model/pipeline.proto



<a name="model-FieldPrefixes"></a>

### FieldPrefixes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inclusion | [bool](#bool) |  | If true, only fields with the specified prefixes will be promoted If false, only fields without the specified prefixes will be promoted |
| prefixes | [string](#string) | repeated |  |






<a name="model-Phase"></a>

### Phase



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_id | [string](#string) |  | Environment to promote to. The project must be active in this environment. |
| field_prefixes | [FieldPrefixes](#model-FieldPrefixes) |  | Fields prefixes to either promote or not. |
| triggers | [Triggers](#model-Triggers) |  | Promotion triggers. |






<a name="model-Pipeline"></a>

### Pipeline



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Custom name for the pipeline. |
| initial_environment | [string](#string) |  | The environment to base the pipeline on. |
| phases | [Phase](#model-Phase) | repeated | The subsequent phases of the pipeline to promote to. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The time the pipeline was created. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The time the pipeline was updated. |
| description | [string](#string) |  | User specified description of the pipeline. |






<a name="model-Trigger"></a>

### Trigger



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| conditions | [Trigger.Condition](#model-Trigger-Condition) | repeated | The conditions that must be met for the trigger to fire. |
| require_all | [bool](#bool) |  | If true, all conditions must be met for the trigger to fire. Otherwise only a single condition must be met. |
| enabled | [bool](#bool) |  | If true, the trigger is enabled. Otherwise it is disabled. |






<a name="model-Trigger-Condition"></a>

### Trigger.Condition
Condition that must be met for the trigger to fire.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| time_alive | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="model-Triggers"></a>

### Triggers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| automatic | [Trigger](#model-Trigger) |  | The automatic trigger |
| manual | [Trigger](#model-Trigger) |  | The manual trigger |













<a name="api_v1_capsule_pipeline_status-proto"></a>

## api/v1/capsule/pipeline/status.proto



<a name="api-v1-capsule-pipeline-PhaseMessage"></a>

### PhaseMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="api-v1-capsule-pipeline-PhaseStatus"></a>

### PhaseStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_id | [string](#string) |  |  |
| state | [PhaseState](#api-v1-capsule-pipeline-PhaseState) |  |  |
| rollout_id | [uint64](#uint64) |  |  |
| messages | [PhaseMessage](#api-v1-capsule-pipeline-PhaseMessage) | repeated |  |
| started_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="api-v1-capsule-pipeline-Status"></a>

### Status



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline_name | [string](#string) |  | The name of the pipeline. |
| capsule_id | [string](#string) |  | The capsule that is executing the pipeline. |
| execution_id | [uint64](#uint64) |  | The ID of the pipeline execution |
| state | [State](#api-v1-capsule-pipeline-State) |  | The overall state of the pipeline execution. |
| phase_statuses | [PhaseStatus](#api-v1-capsule-pipeline-PhaseStatus) | repeated | The statuses of the phases in the pipeline. |
| started_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the pipeline was started. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the pipeline was last updated. |
| current_phase | [uint32](#uint32) |  | current phase |
| config | [model.Pipeline](#model-Pipeline) |  | the configured pipeline |








<a name="api-v1-capsule-pipeline-PhaseState"></a>

### PhaseState


| Name | Number | Description |
| ---- | ------ | ----------- |
| PHASE_STATE_UNSPECIFIED | 0 | The state is unspecified. |
| PHASE_STATE_NOT_READY | 1 | The phase is not ready for promotion |
| PHASE_STATE_READY | 2 | The phase is ready for promotion |
| PHASE_STATE_PROMOTED | 3 | The phase is promoted |



<a name="api-v1-capsule-pipeline-State"></a>

### State


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATE_UNSPECIFIED | 0 | The state is unspecified. |
| STATE_RUNNING | 1 | The pipeline has started. |
| STATE_ABORTED | 2 | The pipeline is aborted. |
| STATE_COMPLETED | 3 | The pipeline is completed. |








<a name="model_revision-proto"></a>

## model/revision.proto



<a name="model-BookmarkingConfiguration"></a>

### BookmarkingConfiguration



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dont_bookmark | [bool](#bool) |  |  |






<a name="model-Fingerprint"></a>

### Fingerprint



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [string](#string) |  |  |






<a name="model-Fingerprints"></a>

### Fingerprints



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [Fingerprint](#model-Fingerprint) |  |  |
| environment | [Fingerprint](#model-Fingerprint) |  |  |
| capsule_set | [Fingerprint](#model-Fingerprint) |  |  |
| capsule | [Fingerprint](#model-Fingerprint) |  |  |






<a name="model-GitLabProposal"></a>

### GitLabProposal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pr_id | [int64](#int64) |  |  |






<a name="model-GithubProposal"></a>

### GithubProposal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pr_id | [int64](#int64) |  |  |






<a name="model-ProposalMetadata"></a>

### ProposalMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| created_by | [Author](#model-Author) |  |  |
| fingerprint | [Fingerprint](#model-Fingerprint) |  |  |
| spawn_point | [RepoBranch](#model-RepoBranch) |  |  |
| branch | [string](#string) |  |  |
| review_url | [string](#string) |  |  |
| github | [GithubProposal](#model-GithubProposal) |  |  |
| gitlab | [GitLabProposal](#model-GitLabProposal) |  |  |






<a name="model-RevisionMetadata"></a>

### RevisionMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| updated_by | [Author](#model-Author) |  |  |
| fingerprint | [Fingerprint](#model-Fingerprint) |  |  |
| git_change | [GitChange](#model-GitChange) |  |  |
| bookmarking | [BookmarkingConfiguration](#model-BookmarkingConfiguration) |  |  |






<a name="model-Revisions"></a>

### Revisions



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| projects | [RevisionMetadata](#model-RevisionMetadata) | repeated |  |
| environments | [RevisionMetadata](#model-RevisionMetadata) | repeated |  |
| capsule_sets | [RevisionMetadata](#model-RevisionMetadata) | repeated |  |
| capsules | [RevisionMetadata](#model-RevisionMetadata) | repeated |  |













<a name="api_v1_capsule_revision-proto"></a>

## api/v1/capsule/revision.proto



<a name="api-v1-capsule-Proposal"></a>

### Proposal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| spec | [platform.v1.Capsule](#platform-v1-Capsule) |  |  |
| metadata | [model.ProposalMetadata](#model-ProposalMetadata) |  |  |






<a name="api-v1-capsule-Revision"></a>

### Revision



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| spec | [platform.v1.Capsule](#platform-v1-Capsule) |  |  |
| metadata | [model.RevisionMetadata](#model-RevisionMetadata) |  |  |
| message | [string](#string) |  |  |






<a name="api-v1-capsule-SetProposal"></a>

### SetProposal



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| spec | [platform.v1.CapsuleSet](#platform-v1-CapsuleSet) |  |  |
| metadata | [model.ProposalMetadata](#model-ProposalMetadata) |  |  |






<a name="api-v1-capsule-SetRevision"></a>

### SetRevision



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| spec | [platform.v1.CapsuleSet](#platform-v1-CapsuleSet) |  |  |
| metadata | [model.RevisionMetadata](#model-RevisionMetadata) |  |  |













<a name="api_v1_capsule_rollout-proto"></a>

## api/v1/capsule/rollout.proto



<a name="api-v1-capsule-Changelog"></a>

### Changelog



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| changes | [Changelog.Change](#api-v1-capsule-Changelog-Change) | repeated |  |
| message | [string](#string) |  |  |






<a name="api-v1-capsule-Changelog-Change"></a>

### Changelog.Change



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |






<a name="api-v1-capsule-ConfigFile"></a>

### ConfigFile



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |
| content | [bytes](#bytes) |  |  |
| updated_by | [model.Author](#model-Author) |  |  |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| is_secret | [bool](#bool) |  |  |






<a name="api-v1-capsule-Rollout"></a>

### Rollout
The rollout model.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rollout_id | [uint64](#uint64) |  | Unique indentifier for the rollout. |
| config | [RolloutConfig](#api-v1-capsule-RolloutConfig) |  | The rollout config. |
| status | [rollout.Status](#api-v1-capsule-rollout-Status) |  | The rollout status. |
| spec | [platform.v1.CapsuleSpec](#platform-v1-CapsuleSpec) |  |  |
| revisions | [model.Revisions](#model-Revisions) |  |  |
| changelog | [Changelog](#api-v1-capsule-Changelog) |  |  |






<a name="api-v1-capsule-RolloutConfig"></a>

### RolloutConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created_by | [model.Author](#model-Author) |  | The user who initiated the rollout. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| replicas | [uint32](#uint32) |  |  |
| image_id | [string](#string) |  |  |
| network | [Network](#api-v1-capsule-Network) |  |  |
| container_settings | [ContainerSettings](#api-v1-capsule-ContainerSettings) |  |  |
| auto_add_rig_service_accounts | [bool](#bool) |  |  |
| config_files | [ConfigFile](#api-v1-capsule-ConfigFile) | repeated |  |
| horizontal_scale | [HorizontalScale](#api-v1-capsule-HorizontalScale) |  |  |
| cron_jobs | [CronJob](#api-v1-capsule-CronJob) | repeated |  |
| environment_id | [string](#string) |  |  |
| message | [string](#string) |  |  |
| annotations | [RolloutConfig.AnnotationsEntry](#api-v1-capsule-RolloutConfig-AnnotationsEntry) | repeated |  |






<a name="api-v1-capsule-RolloutConfig-AnnotationsEntry"></a>

### RolloutConfig.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |








<a name="api-v1-capsule-EventType"></a>

### EventType


| Name | Number | Description |
| ---- | ------ | ----------- |
| EVENT_TYPE_UNSPECIFIED | 0 |  |
| EVENT_TYPE_ABORT | 1 |  |








<a name="api_v1_capsule_status-proto"></a>

## api/v1/capsule/status.proto



<a name="api-v1-capsule-CapsuleStatus"></a>

### CapsuleStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| statuses | [api.v1.pipeline.ObjectStatus](#api-v1-pipeline-ObjectStatus) | repeated |  |






<a name="api-v1-capsule-ConfigFileStatus"></a>

### ConfigFileStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |
| isSecret | [bool](#bool) |  |  |
| status | [api.v1.pipeline.ObjectStatus](#api-v1-pipeline-ObjectStatus) | repeated |  |
| transition | [Transition](#api-v1-capsule-Transition) |  |  |
| content | [bytes](#bytes) |  |  |






<a name="api-v1-capsule-ContainerConfig"></a>

### ContainerConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| image | [string](#string) |  |  |
| command | [string](#string) |  |  |
| args | [string](#string) | repeated |  |
| environment_variables | [ContainerConfig.EnvironmentVariablesEntry](#api-v1-capsule-ContainerConfig-EnvironmentVariablesEntry) | repeated |  |
| scale | [HorizontalScale](#api-v1-capsule-HorizontalScale) |  |  |
| resources | [Resources](#api-v1-capsule-Resources) |  |  |






<a name="api-v1-capsule-ContainerConfig-EnvironmentVariablesEntry"></a>

### ContainerConfig.EnvironmentVariablesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="api-v1-capsule-CronJobStatus"></a>

### CronJobStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job_name | [string](#string) |  |  |
| schedule | [string](#string) |  |  |
| last_execution | [api.v1.pipeline.ObjectState](#api-v1-pipeline-ObjectState) |  |  |
| transition | [Transition](#api-v1-capsule-Transition) |  |  |






<a name="api-v1-capsule-InstancesStatus"></a>

### InstancesStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| num_upgrading | [uint32](#uint32) |  | The number of updated instances. |
| num_ready | [uint32](#uint32) |  | The number of ready instances. |
| num_stuck | [uint32](#uint32) |  | The number of stuck instances. |
| num_wrong_version | [uint32](#uint32) |  | The number of instances with the wrong version. |






<a name="api-v1-capsule-InterfaceStatus"></a>

### InterfaceStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| port | [uint32](#uint32) |  |  |
| routes | [InterfaceStatus.Route](#api-v1-capsule-InterfaceStatus-Route) | repeated |  |
| status | [api.v1.pipeline.ObjectStatus](#api-v1-pipeline-ObjectStatus) | repeated |  |
| transition | [Transition](#api-v1-capsule-Transition) |  |  |






<a name="api-v1-capsule-InterfaceStatus-Route"></a>

### InterfaceStatus.Route



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| route | [HostRoute](#api-v1-capsule-HostRoute) |  |  |
| status | [api.v1.pipeline.ObjectStatus](#api-v1-pipeline-ObjectStatus) | repeated |  |
| transition | [Transition](#api-v1-capsule-Transition) |  |  |






<a name="api-v1-capsule-Status"></a>

### Status



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespace | [string](#string) |  |  |
| capsule | [CapsuleStatus](#api-v1-capsule-CapsuleStatus) |  |  |
| current_rollout_id | [uint64](#uint64) |  |  |
| container_config | [ContainerConfig](#api-v1-capsule-ContainerConfig) |  |  |
| instances | [InstancesStatus](#api-v1-capsule-InstancesStatus) |  |  |
| interfaces | [InterfaceStatus](#api-v1-capsule-InterfaceStatus) | repeated |  |
| config_files | [ConfigFileStatus](#api-v1-capsule-ConfigFileStatus) | repeated |  |
| cron_jobs | [CronJobStatus](#api-v1-capsule-CronJobStatus) | repeated |  |
| issues | [model.Issue](#model-Issue) | repeated | List of all issues associated to the Capsule, include those of the current rollout. The list does not include instance-level issues. |








<a name="api-v1-capsule-Transition"></a>

### Transition


| Name | Number | Description |
| ---- | ------ | ----------- |
| TRANSITION_UNSPECIFIED | 0 |  |
| TRANSITION_BEING_CREATED | 1 |  |
| TRANSITION_UP_TO_DATE | 2 |  |
| TRANSITION_BEING_DELETED | 3 |  |








<a name="api_v1_capsule_service-proto"></a>

## api/v1/capsule/service.proto



<a name="api-v1-capsule-AbortPipelineRequest"></a>

### AbortPipelineRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| execution_id | [uint64](#uint64) |  |  |






<a name="api-v1-capsule-AbortPipelineResponse"></a>

### AbortPipelineResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [pipeline.Status](#api-v1-capsule-pipeline-Status) |  |  |






<a name="api-v1-capsule-AbortRolloutRequest"></a>

### AbortRolloutRequest
AbortRolloutRequest aborts a rollout.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to abort the rollout in. |
| rollout_id | [uint64](#uint64) |  | The rollout to abort. |
| project_id | [string](#string) |  | The project in which the capsule lives. |






<a name="api-v1-capsule-AbortRolloutResponse"></a>

### AbortRolloutResponse
AbortRolloutResponse is an empty response.






<a name="api-v1-capsule-CapsuleMetricsRequest"></a>

### CapsuleMetricsRequest
Request for getting metrics for a capsule and optionally a single instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to get metrics for. |
| instance_id | [string](#string) |  | If set, only returns metrics for the given instance_id. |
| project_id | [string](#string) |  | The project in which the capsule lives. |
| environment_id | [string](#string) |  | The environment to get metrics for. |
| since | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Return metrics generated after 'since' |






<a name="api-v1-capsule-CapsuleMetricsResponse"></a>

### CapsuleMetricsResponse
Response to getting capsule metrics.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instance_metrics | [model.InstanceMetrics](#model-InstanceMetrics) | repeated | Metrics |






<a name="api-v1-capsule-CreateRequest"></a>

### CreateRequest
Create capsule request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | The name of the capsule. This property must be unique for a project and cannot be changed after creation. Resources created in associating with the capsule will use this name. |
| initializers | [Update](#api-v1-capsule-Update) | repeated | Deprecated field: The initial properties of the capsule. |
| project_id | [string](#string) |  | The project to create the capsule in. |






<a name="api-v1-capsule-CreateResponse"></a>

### CreateResponse
Create capsule response.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | ID of the capsule. This is the same as the name. |






<a name="api-v1-capsule-DeleteRequest"></a>

### DeleteRequest
Request to delete a capsule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to delete. |
| project_id | [string](#string) |  | The project in which the capsule is to be deleted. |






<a name="api-v1-capsule-DeleteResponse"></a>

### DeleteResponse
Empty delete response.






<a name="api-v1-capsule-DeployOutcome"></a>

### DeployOutcome



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| field_changes | [FieldChange](#api-v1-capsule-FieldChange) | repeated | The field-level changes that comes from applying this change. |
| platform_objects | [DeployOutcome.PlatformObject](#api-v1-capsule-DeployOutcome-PlatformObject) | repeated | The Platform-level objects that are generated by the Deploy. |
| kubernetes_objects | [DeployOutcome.KubernetesObject](#api-v1-capsule-DeployOutcome-KubernetesObject) | repeated | The Kubernetes-level objects that are generated by the Deploy. The objects are both the outcome of what the platform generated for the Kubernetes cluster directly and what plugins are further adding. |
| kubernetes_error | [string](#string) |  | In case generation of kubernetes files failed, this field will be populated with the error. |
| cluster_name | [string](#string) |  |  |
| kubernetes_namespace | [string](#string) |  |  |






<a name="api-v1-capsule-DeployOutcome-KubernetesObject"></a>

### DeployOutcome.KubernetesObject



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| content_yaml | [string](#string) |  |  |






<a name="api-v1-capsule-DeployOutcome-PlatformObject"></a>

### DeployOutcome.PlatformObject



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| content_yaml | [string](#string) |  |  |






<a name="api-v1-capsule-DeployRequest"></a>

### DeployRequest
Deploy request. This will deploy a number of changes which results in a new
rollout.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule to deploy to. |
| changes | [Change](#api-v1-capsule-Change) | repeated | Changes to include in the new rollout. |
| force | [bool](#bool) |  | Force deploy, aborting an existing rollout if ongoing. |
| project_id | [string](#string) |  | Project in which the capsule lives. |
| environment_id | [string](#string) |  | Environment in which to deploy. |
| message | [string](#string) |  | Deploy message. |
| dry_run | [bool](#bool) |  | if true, the deploy will not be executed, but the request will return the rollout config. |
| current_rollout_id | [uint64](#uint64) |  | If not zero, this will constrain the rollout only to be created if the currently running rollout matches this identifier. If this check fails, the request will return an `Aborted` error. |
| current_fingerprint | [model.Fingerprint](#model-Fingerprint) |  | If set, this will constrain the rollout only to be created if the current latest capsule fingerprint matches the given. Cannot be used together with `current_rollout_id` |
| force_override | [bool](#bool) |  | By default, existing objects will be kept in favor of overriding them. To force the override of resources, set this flag to true. An example of this use-case is a migration step, where resource created by a previous toolchain e.g. based on Helm charts, are to be replaced and instead be created by the Rig operator. While the override is irreversible, this flag is not "sticky" and must be set by each deploy that should use this behavior. |
| operator_config | [string](#string) |  | Only allowed with dry_run = true. Will use this config for the operator instead of the config running in the cluster. |






<a name="api-v1-capsule-DeployResponse"></a>

### DeployResponse
Deploy response.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rollout_id | [uint64](#uint64) |  | ID of the new rollout. |
| resource_yaml | [DeployResponse.ResourceYamlEntry](#api-v1-capsule-DeployResponse-ResourceYamlEntry) | repeated | The YAML of the resources that will be deployed. Deprecated. Use `outcome` instead. |
| revision | [Revision](#api-v1-capsule-Revision) |  | The rollout config. api.v1.capsule.RolloutConfig rollout_config = 3; The capsule revision created. |
| set_revision | [SetRevision](#api-v1-capsule-SetRevision) |  | The capsule set revision created if it's the first time deploying to the environment. |
| outcome | [DeployOutcome](#api-v1-capsule-DeployOutcome) |  | Breakdown of the changes that this deploy would make to the system. Only populated if dry-run is used. |






<a name="api-v1-capsule-DeployResponse-ResourceYamlEntry"></a>

### DeployResponse.ResourceYamlEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="api-v1-capsule-DeploySetOutcome"></a>

### DeploySetOutcome



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| field_changes | [FieldChange](#api-v1-capsule-FieldChange) | repeated | The field-level changes that comes from applying this change. |
| environments | [DeploySetOutcome.EnvironmentsEntry](#api-v1-capsule-DeploySetOutcome-EnvironmentsEntry) | repeated |  |






<a name="api-v1-capsule-DeploySetOutcome-EnvironmentsEntry"></a>

### DeploySetOutcome.EnvironmentsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [DeployOutcome](#api-v1-capsule-DeployOutcome) |  |  |






<a name="api-v1-capsule-DeploySetRequest"></a>

### DeploySetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule to deploy to. |
| changes | [Change](#api-v1-capsule-Change) | repeated | Changes to include in the new rollout. |
| force | [bool](#bool) |  | Force deploy, aborting existing rollouts if ongoing. |
| project_id | [string](#string) |  | Project in which the capsule lives. |
| message | [string](#string) |  | Deploy message. |
| dry_run | [bool](#bool) |  | if true, the deploy will not be executed, but the request will return the rollout config. |
| current_rollout_ids | [DeploySetRequest.CurrentRolloutIdsEntry](#api-v1-capsule-DeploySetRequest-CurrentRolloutIdsEntry) | repeated | If present, maps from environment to expected current rollout within that environment. This will constrain the rollout only to be created if the currently running rollout matches this identifier. If this check fails, the request will return an `Aborted` error. |
| current_fingerprint | [model.Fingerprint](#model-Fingerprint) |  | If set, this will constrain the rollout only to be created if the current latest capsuleset fingerprint matches the given. |
| current_environment_fingerprints | [DeploySetRequest.CurrentEnvironmentFingerprintsEntry](#api-v1-capsule-DeploySetRequest-CurrentEnvironmentFingerprintsEntry) | repeated | If set, this will constrain the rollout only to be created if the current latest capsule fingerprint for each environment in the map matches the ones in the map. Cannot be used together with `current_rollout_ids` |






<a name="api-v1-capsule-DeploySetRequest-CurrentEnvironmentFingerprintsEntry"></a>

### DeploySetRequest.CurrentEnvironmentFingerprintsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [model.Fingerprint](#model-Fingerprint) |  |  |






<a name="api-v1-capsule-DeploySetRequest-CurrentRolloutIdsEntry"></a>

### DeploySetRequest.CurrentRolloutIdsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [uint64](#uint64) |  |  |






<a name="api-v1-capsule-DeploySetResponse"></a>

### DeploySetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| revision | [SetRevision](#api-v1-capsule-SetRevision) |  | The capsule revision created. |
| outcome | [DeploySetOutcome](#api-v1-capsule-DeploySetOutcome) |  | Breakdown of the changes that this deploy would make to the system. Only populated if dry-run is used. |
| ActiveEnvironments | [string](#string) | repeated | The environments which currently have rollouts. These will receive a rollout as result of the SetDeploy |






<a name="api-v1-capsule-ExecuteRequest"></a>

### ExecuteRequest
Execute request. This can either be a request to start a request, a terminal
resize msg or a stream data msg.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start | [ExecuteRequest.Start](#api-v1-capsule-ExecuteRequest-Start) |  | Start request. |
| stdin | [StreamData](#api-v1-capsule-StreamData) |  | Stream stdin request |
| resize | [ExecuteRequest.Resize](#api-v1-capsule-ExecuteRequest-Resize) |  | Resize request |
| project_id | [string](#string) |  | The project ID. |
| environment_id | [string](#string) |  | The environment ID. |






<a name="api-v1-capsule-ExecuteRequest-Resize"></a>

### ExecuteRequest.Resize
Terminal resize request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| height | [uint32](#uint32) |  | The new terminal height. |
| width | [uint32](#uint32) |  | The new terminal width. |






<a name="api-v1-capsule-ExecuteRequest-Start"></a>

### ExecuteRequest.Start
Exec start request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to execute in. |
| instance_id | [string](#string) |  | The instance to execute in. |
| command | [string](#string) |  | The command to execute. |
| arguments | [string](#string) | repeated | The arguments to the command. |
| tty | [ExecuteRequest.Resize](#api-v1-capsule-ExecuteRequest-Resize) |  | The initial terminal size. |
| interactive | [bool](#bool) |  | If the command is interactive. |






<a name="api-v1-capsule-ExecuteResponse"></a>

### ExecuteResponse
Execute response.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| stdout | [StreamData](#api-v1-capsule-StreamData) |  | Stdout of the execute. |
| stderr | [StreamData](#api-v1-capsule-StreamData) |  | Stderr in case of an error. |
| exit_code | [int32](#int32) |  | Exit code of the execute. |






<a name="api-v1-capsule-GetCustomInstanceMetricsRequest"></a>

### GetCustomInstanceMetricsRequest
Request for getting custom metrics for a capsule in an environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to get metrics for. |
| project_id | [string](#string) |  | The project in which the capsule lives. |
| environment_id | [string](#string) |  | The environment to get metrics for. |






<a name="api-v1-capsule-GetCustomInstanceMetricsResponse"></a>

### GetCustomInstanceMetricsResponse
Response to getting custom metrics for a capsule in an environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metrics | [model.Metric](#model-Metric) | repeated | Custom Metrics. |






<a name="api-v1-capsule-GetEffectiveGitSettingsRequest"></a>

### GetEffectiveGitSettingsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| environment_id | [string](#string) |  |  |
| capsule_id | [string](#string) |  |  |






<a name="api-v1-capsule-GetEffectiveGitSettingsResponse"></a>

### GetEffectiveGitSettingsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| git | [model.GitStore](#model-GitStore) |  |  |
| environment_enabled | [bool](#bool) |  |  |






<a name="api-v1-capsule-GetInstanceStatusRequest"></a>

### GetInstanceStatusRequest
Get status of an instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to get the instance status from. |
| instance_id | [string](#string) |  | The instance to get. |
| project_id | [string](#string) |  | The project in which the capsule lives. |
| environment_id | [string](#string) |  | The environment to get the instance from. |






<a name="api-v1-capsule-GetInstanceStatusResponse"></a>

### GetInstanceStatusResponse
Get instance status response.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [instance.Status](#api-v1-capsule-instance-Status) |  | The instance status. |






<a name="api-v1-capsule-GetJobExecutionsRequest"></a>

### GetJobExecutionsRequest
Request for getting job executions from cron jobs.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to get job executions for. |
| job_name | [string](#string) |  | The name of the job to get executions for. |
| states | [JobState](#api-v1-capsule-JobState) | repeated | Filtering executions by job state. |
| created_from | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Filtering executions created before this timestamp. |
| created_to | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Filtering executions created after this timestamp. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |
| project_id | [string](#string) |  | The project in which the capsule lives. |
| environment_id | [string](#string) |  | The environment to get job executions for. |






<a name="api-v1-capsule-GetJobExecutionsResponse"></a>

### GetJobExecutionsResponse
Response to getting job executions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job_executions | [JobExecution](#api-v1-capsule-JobExecution) | repeated | Job executions. |
| total | [uint64](#uint64) |  | Total number of executions ignorring pagination. |






<a name="api-v1-capsule-GetPipelineStatusRequest"></a>

### GetPipelineStatusRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| execution_id | [uint64](#uint64) |  |  |






<a name="api-v1-capsule-GetPipelineStatusResponse"></a>

### GetPipelineStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [pipeline.Status](#api-v1-capsule-pipeline-Status) |  |  |






<a name="api-v1-capsule-GetProposalsEnabledRequest"></a>

### GetProposalsEnabledRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| environment_id | [string](#string) |  |  |
| capsule_id | [string](#string) |  |  |






<a name="api-v1-capsule-GetProposalsEnabledResponse"></a>

### GetProposalsEnabledResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  |  |






<a name="api-v1-capsule-GetRequest"></a>

### GetRequest
Request to get a capsule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule to get. |
| project_id | [string](#string) |  | Project in which the capsule is. |






<a name="api-v1-capsule-GetResponse"></a>

### GetResponse
Response to get a capsule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule | [Capsule](#api-v1-capsule-Capsule) |  | The capsule. |
| revision | [SetRevision](#api-v1-capsule-SetRevision) |  |  |
| environment_revisions | [GetResponse.EnvironmentRevisionsEntry](#api-v1-capsule-GetResponse-EnvironmentRevisionsEntry) | repeated |  |






<a name="api-v1-capsule-GetResponse-EnvironmentRevisionsEntry"></a>

### GetResponse.EnvironmentRevisionsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Revision](#api-v1-capsule-Revision) |  |  |






<a name="api-v1-capsule-GetRevisionRequest"></a>

### GetRevisionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| environment_id | [string](#string) |  |  |
| capsule_id | [string](#string) |  |  |
| fingerprint | [model.Fingerprint](#model-Fingerprint) |  |  |






<a name="api-v1-capsule-GetRevisionResponse"></a>

### GetRevisionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| revision | [Revision](#api-v1-capsule-Revision) |  |  |






<a name="api-v1-capsule-GetRolloutOfRevisionsRequest"></a>

### GetRolloutOfRevisionsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| environment_id | [string](#string) |  |  |
| capsule_id | [string](#string) |  |  |
| fingerprints | [model.Fingerprints](#model-Fingerprints) |  |  |






<a name="api-v1-capsule-GetRolloutOfRevisionsResponse"></a>

### GetRolloutOfRevisionsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| no_rollout | [GetRolloutOfRevisionsResponse.NoRollout](#api-v1-capsule-GetRolloutOfRevisionsResponse-NoRollout) |  |  |
| rollout | [Rollout](#api-v1-capsule-Rollout) |  |  |






<a name="api-v1-capsule-GetRolloutOfRevisionsResponse-NoRollout"></a>

### GetRolloutOfRevisionsResponse.NoRollout



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [bool](#bool) |  | Indicates no rollout with a project revision at least as new as the one given. |
| environment | [bool](#bool) |  | Indicates no rollout with an environment revision at least as new as the one given. |
| capsule_set | [bool](#bool) |  | Indicates no rollout with a capsule set revision at least as new as the one given. |
| capsule | [bool](#bool) |  | Indicates no rollout with a capsule revision at least as new as the one given. |






<a name="api-v1-capsule-GetRolloutRequest"></a>

### GetRolloutRequest
GetRolloutRequest gets a single rollout.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to get the rollout from. |
| rollout_id | [uint64](#uint64) |  | The rollout to get. |
| project_id | [string](#string) |  | The project in which the capsule lives. |






<a name="api-v1-capsule-GetRolloutResponse"></a>

### GetRolloutResponse
GetRolloutResponse returns a single rollout for a capsule and an environment
in a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rollout | [Rollout](#api-v1-capsule-Rollout) |  | The rollout. |






<a name="api-v1-capsule-GetStatusRequest"></a>

### GetStatusRequest
GetStatusRequest is a request to start streaming the capsule status


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to get the status from. |
| project_id | [string](#string) |  | The project in which the capsule lives. |
| environment_id | [string](#string) |  | The environment to get the status from. |






<a name="api-v1-capsule-GetStatusResponse"></a>

### GetStatusResponse
GetCapsuleStatusResponse


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [Status](#api-v1-capsule-Status) |  | The capsule status |






<a name="api-v1-capsule-ListEventsRequest"></a>

### ListEventsRequest
ListEvents request for listing rollout events for a given rollout in a
capsule and environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to list events for. |
| rollout_id | [uint64](#uint64) |  | The rollout to list events for. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |
| project_id | [string](#string) |  | The project in which the capsule lives. |
| environment_id | [string](#string) |  | The environment to list events for. |






<a name="api-v1-capsule-ListEventsResponse"></a>

### ListEventsResponse
Response to List Events


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| events | [Event](#api-v1-capsule-Event) | repeated | The events for a rollout in a capsule and environment for a given project. |
| total | [uint64](#uint64) |  | Total number of events in the capsule for the given environment. |






<a name="api-v1-capsule-ListInstanceStatusesRequest"></a>

### ListInstanceStatusesRequest
List multiple instance statuses


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to get the instance statuses from. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |
| project_id | [string](#string) |  | The project in which the capsule is. |
| environment_id | [string](#string) |  | The environment to get the instance statuses from. |
| include_deleted | [bool](#bool) |  | if true, deleted instances will be included in the response. |
| exclude_existing | [bool](#bool) |  | if true, existing instances will be excluded from the response. |






<a name="api-v1-capsule-ListInstanceStatusesResponse"></a>

### ListInstanceStatusesResponse
Response for listing multiple instance statuses


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instances | [instance.Status](#api-v1-capsule-instance-Status) | repeated | The instance statuses. |
| total | [uint64](#uint64) |  | Total number of instances in the capsule for the given environment. |






<a name="api-v1-capsule-ListInstancesRequest"></a>

### ListInstancesRequest
List instances request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule to list instances from. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |
| project_id | [string](#string) |  | Project in which the capsule lives. |
| environment_id | [string](#string) |  | Environment to list instances from. |
| include_deleted | [bool](#bool) |  | if true, deleted instances will be included in the response. |
| exclude_existing | [bool](#bool) |  | if true, existing instances will be excluded from the response. |






<a name="api-v1-capsule-ListInstancesResponse"></a>

### ListInstancesResponse
List instances response.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| instances | [Instance](#api-v1-capsule-Instance) | repeated | The instances. |
| total | [uint64](#uint64) |  | Total number of instances in the capsule for the given environment. |






<a name="api-v1-capsule-ListPipelineStatusesRequest"></a>

### ListPipelineStatusesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [model.Pagination](#model-Pagination) |  |  |
| project_filter | [string](#string) |  | Only include pipelines that are run in the given project. |
| capsule_filter | [string](#string) |  | Only include pipelines that are run with the given capsule. |
| states_filter | [pipeline.State](#api-v1-capsule-pipeline-State) | repeated | Only include pipelines that are in one of the given states. |
| name_filter | [string](#string) |  | Only include pipelines that have the given name. |






<a name="api-v1-capsule-ListPipelineStatusesResponse"></a>

### ListPipelineStatusesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| statuses | [pipeline.Status](#api-v1-capsule-pipeline-Status) | repeated |  |






<a name="api-v1-capsule-ListProposalsRequest"></a>

### ListProposalsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| environment_id | [string](#string) |  |  |
| capsule_id | [string](#string) |  |  |
| pagination | [model.Pagination](#model-Pagination) |  |  |






<a name="api-v1-capsule-ListProposalsResponse"></a>

### ListProposalsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| proposals | [Proposal](#api-v1-capsule-Proposal) | repeated |  |
| total | [uint64](#uint64) |  |  |






<a name="api-v1-capsule-ListRequest"></a>

### ListRequest
List capsule request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |
| project_id | [string](#string) |  | Project in which to list capsules. |






<a name="api-v1-capsule-ListResponse"></a>

### ListResponse
List capsule response.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsules | [Capsule](#api-v1-capsule-Capsule) | repeated | The capsules. |
| total | [uint64](#uint64) |  | Total number of capsules in the project. |






<a name="api-v1-capsule-ListRolloutsRequest"></a>

### ListRolloutsRequest
ListRolloutsRequest lists rollouts for a capsule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to list rollouts for. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |
| project_id | [string](#string) |  | The project in which the capsule lives. |
| environment_id | [string](#string) |  | The environment to list rollouts for. |






<a name="api-v1-capsule-ListRolloutsResponse"></a>

### ListRolloutsResponse
ListRolloutsResponse lists rollouts for a capsule and an environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rollouts | [Rollout](#api-v1-capsule-Rollout) | repeated | The rollouts. |
| total | [uint64](#uint64) |  | Total number of rollouts in the capsule for the given environment. |






<a name="api-v1-capsule-ListSetProposalsRequest"></a>

### ListSetProposalsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| capsule_id | [string](#string) |  |  |
| pagination | [model.Pagination](#model-Pagination) |  |  |






<a name="api-v1-capsule-ListSetProposalsResponse"></a>

### ListSetProposalsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| proposals | [SetProposal](#api-v1-capsule-SetProposal) | repeated |  |
| total | [uint64](#uint64) |  |  |






<a name="api-v1-capsule-LogsRequest"></a>

### LogsRequest
Request to get instance logs from a capsule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to read logs from. |
| instance_id | [string](#string) |  | The instance in the capsule to read logs from. |
| follow | [bool](#bool) |  | If true, the request will stay open and stream new log messages. |
| since | [google.protobuf.Duration](#google-protobuf-Duration) |  | If set, will not show logs older than since. |
| project_id | [string](#string) |  | The project in which the capsule is. |
| environment_id | [string](#string) |  | Environment to get logs from. |
| previous_containers | [bool](#bool) |  | If true, include logs from previously terminated containers |






<a name="api-v1-capsule-LogsResponse"></a>

### LogsResponse
The response of a capsule.Logs RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| log | [Log](#api-v1-capsule-Log) |  | The actual logs |






<a name="api-v1-capsule-PipelineDryRunOutput"></a>

### PipelineDryRunOutput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_id | [string](#string) |  | Environment to promote to. |
| outcome | [DeployOutcome](#api-v1-capsule-DeployOutcome) |  | Breakdown of the changes that this deploy would make to the system. |
| revision | [Revision](#api-v1-capsule-Revision) |  |  |






<a name="api-v1-capsule-PortForwardRequest"></a>

### PortForwardRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start | [PortForwardRequest.Start](#api-v1-capsule-PortForwardRequest-Start) |  |  |
| data | [bytes](#bytes) |  |  |
| close | [PortForwardRequest.Close](#api-v1-capsule-PortForwardRequest-Close) |  |  |






<a name="api-v1-capsule-PortForwardRequest-Close"></a>

### PortForwardRequest.Close







<a name="api-v1-capsule-PortForwardRequest-Start"></a>

### PortForwardRequest.Start



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  | The project ID. |
| environment_id | [string](#string) |  | The environment ID. |
| capsule_id | [string](#string) |  |  |
| instance_id | [string](#string) |  |  |
| port | [uint32](#uint32) |  |  |






<a name="api-v1-capsule-PortForwardResponse"></a>

### PortForwardResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [bytes](#bytes) |  |  |
| close | [PortForwardResponse.Close](#api-v1-capsule-PortForwardResponse-Close) |  |  |






<a name="api-v1-capsule-PortForwardResponse-Close"></a>

### PortForwardResponse.Close







<a name="api-v1-capsule-PromotePipelineRequest"></a>

### PromotePipelineRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| execution_id | [uint64](#uint64) |  |  |
| dry_run | [bool](#bool) |  | If true, the progression will not be executed, but instead a breakdown of changes will be returned |
| field_changes | [FieldChange](#api-v1-capsule-FieldChange) | repeated | additional changes to include in the manual promotion |
| force | [bool](#bool) |  | If true, the pipeline will be force promoted to the next environment regardless of the state of the pipeline and the triggers. |






<a name="api-v1-capsule-PromotePipelineResponse"></a>

### PromotePipelineResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [pipeline.Status](#api-v1-capsule-pipeline-Status) |  |  |
| dry_run_outcomes | [PipelineDryRunOutput](#api-v1-capsule-PipelineDryRunOutput) | repeated | Breakdown of the changes that will be made to the environments throughout the pipeline. Only populated if dry-run is used. |
| revision | [Revision](#api-v1-capsule-Revision) |  | The capsule revision created. |






<a name="api-v1-capsule-ProposeRolloutRequest"></a>

### ProposeRolloutRequest
Deploy request. This will deploy a number of changes which results in a new
rollout.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule to deploy to. |
| changes | [Change](#api-v1-capsule-Change) | repeated | Changes to include in the new rollout. |
| project_id | [string](#string) |  | Project in which the capsule lives. |
| environment_id | [string](#string) |  | Environment in which to deploy. |
| message | [string](#string) |  | Deploy message. |
| force_override | [bool](#bool) |  | By default, existing objects will be kept in favor of overriding them. To force the override of resources, set this flag to true. An example of this use-case is a migration step, where resource created by a previous toolchain e.g. based on Helm charts, are to be replaced and instead be created by the Rig operator. While the override is irreversible, this flag is not "sticky" and must be set by each deploy that should use this behavior. |
| branch_name | [string](#string) |  |  |






<a name="api-v1-capsule-ProposeRolloutResponse"></a>

### ProposeRolloutResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| proposal | [Proposal](#api-v1-capsule-Proposal) |  |  |
| outcome | [DeployOutcome](#api-v1-capsule-DeployOutcome) |  | Breakdown of the changes that this deploy would make to the system. |






<a name="api-v1-capsule-ProposeSetRolloutRequest"></a>

### ProposeSetRolloutRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule to deploy to. |
| changes | [Change](#api-v1-capsule-Change) | repeated | Changes to include in the new rollout. |
| project_id | [string](#string) |  | Project in which the capsule lives. |
| message | [string](#string) |  | Deploy message. |
| force_override | [bool](#bool) |  | By default, existing objects will be kept in favor of overriding them. To force the override of resources, set this flag to true. An example of this use-case is a migration step, where resource created by a previous toolchain e.g. based on Helm charts, are to be replaced and instead be created by the Rig operator. While the override is irreversible, this flag is not "sticky" and must be set by each deploy that should use this behavior. |
| branch_name | [string](#string) |  |  |






<a name="api-v1-capsule-ProposeSetRolloutResponse"></a>

### ProposeSetRolloutResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| proposal | [SetProposal](#api-v1-capsule-SetProposal) |  |  |
| outcome | [DeploySetOutcome](#api-v1-capsule-DeploySetOutcome) |  | Breakdown of the changes that this deploy would make to the system. |






<a name="api-v1-capsule-RestartInstanceRequest"></a>

### RestartInstanceRequest
RestartInstanceRequest restarts a single instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to restart the instance in. |
| instance_id | [string](#string) |  | The instance to restart. |
| project_id | [string](#string) |  | The project in which the capsule lives. |
| environment_id | [string](#string) |  | The environment to restart the instance in. |






<a name="api-v1-capsule-RestartInstanceResponse"></a>

### RestartInstanceResponse
RestartInstanceResponse is an empty response.






<a name="api-v1-capsule-StartPipelineRequest"></a>

### StartPipelineRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| capsule_id | [string](#string) |  |  |
| pipeline_name | [string](#string) |  |  |
| dry_run | [bool](#bool) |  |  |
| abort_current | [bool](#bool) |  | If true, and the pipeline is already running for the capsule and project, it will be aborted and a new one started. |






<a name="api-v1-capsule-StartPipelineResponse"></a>

### StartPipelineResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [pipeline.Status](#api-v1-capsule-pipeline-Status) |  |  |
| dry_run_outcomes | [PipelineDryRunOutput](#api-v1-capsule-PipelineDryRunOutput) | repeated | Breakdown of the changes that will be made to the environments throughout the pipeline. Only populated if dry-run is used. |






<a name="api-v1-capsule-StopRolloutRequest"></a>

### StopRolloutRequest
StopRolloutRequest aborts a rollout.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule of the aborting rollout. |
| rollout_id | [uint64](#uint64) |  | The rollout to stop. |
| project_id | [string](#string) |  | The project in which the capsule lives. |






<a name="api-v1-capsule-StopRolloutResponse"></a>

### StopRolloutResponse
StopRolloutResponse is an empty response.






<a name="api-v1-capsule-StreamData"></a>

### StreamData
StreamData for Execute RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [bytes](#bytes) |  | Stream data. |
| closed | [bool](#bool) |  | If the stream is closed. |






<a name="api-v1-capsule-UpdateRequest"></a>

### UpdateRequest
Deprecated update - This is now a no-op


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to update. |
| updates | [Update](#api-v1-capsule-Update) | repeated | The updates to apply to the capsule. |
| project_id | [string](#string) |  |  |






<a name="api-v1-capsule-UpdateResponse"></a>

### UpdateResponse
Deprecated: Empty update response.






<a name="api-v1-capsule-WatchInstanceStatusesRequest"></a>

### WatchInstanceStatusesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to get the instance statuses from. |
| project_id | [string](#string) |  | The project in which the capsule is. |
| environment_id | [string](#string) |  | The environment to get the instance statuses from. |
| instance_id | [string](#string) |  | If given, only the instance with this ID will be watched. |
| include_deleted | [bool](#bool) |  | if true, deleted instances will be included in the response. |
| exclude_existing | [bool](#bool) |  | if true, existing instances will be excluded from the response. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






<a name="api-v1-capsule-WatchInstanceStatusesResponse"></a>

### WatchInstanceStatusesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| updated | [instance.Status](#api-v1-capsule-instance-Status) |  |  |
| deleted | [string](#string) |  |  |






<a name="api-v1-capsule-WatchRolloutsRequest"></a>

### WatchRolloutsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to list rollouts for. |
| project_id | [string](#string) |  | The project in which the capsule lives. |
| environment_id | [string](#string) |  | The environment to list rollouts for. |
| rollout_id | [uint64](#uint64) |  | If given only the rollout with this ID will be watched. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






<a name="api-v1-capsule-WatchRolloutsResponse"></a>

### WatchRolloutsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| updated | [Rollout](#api-v1-capsule-Rollout) |  |  |






<a name="api-v1-capsule-WatchStatusRequest"></a>

### WatchStatusRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | The capsule to watch the status of. |
| project_id | [string](#string) |  | The project in which the capsule lives. |
| environment_id | [string](#string) |  | The environment to watch the status of. |






<a name="api-v1-capsule-WatchStatusResponse"></a>

### WatchStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [Status](#api-v1-capsule-Status) |  |  |













<a name="api_v1_capsule_sidecar-proto"></a>

## api/v1/capsule/sidecar.proto



<a name="api-v1-capsule-Sidecar"></a>

### Sidecar
Deprecated: sidecar configuration


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| interfaces | [SidecarInterface](#api-v1-capsule-SidecarInterface) | repeated |  |






<a name="api-v1-capsule-SidecarInterface"></a>

### SidecarInterface
Deprecated: sidecar interface configuration


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [uint32](#uint32) |  |  |
| proxy_port | [uint32](#uint32) |  |  |













<a name="api_v1_cluster_cluster-proto"></a>

## api/v1/cluster/cluster.proto



<a name="api-v1-cluster-Cluster"></a>

### Cluster
Cluster model.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_id | [string](#string) |  | ID of the cluster. |













<a name="api_v1_cluster_service-proto"></a>

## api/v1/cluster/service.proto



<a name="api-v1-cluster-DockerDaemon"></a>

### DockerDaemon
Docker daemon dev registry






<a name="api-v1-cluster-GetConfigRequest"></a>

### GetConfigRequest
request for getting cluster config for an environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_id | [string](#string) |  | The environment to get cluster config for. |






<a name="api-v1-cluster-GetConfigResponse"></a>

### GetConfigResponse
response for getting cluster config for an environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_type | [ClusterType](#api-v1-cluster-ClusterType) |  | Type of the cluster. |
| docker | [DockerDaemon](#api-v1-cluster-DockerDaemon) |  | Docker. |
| registry | [Registry](#api-v1-cluster-Registry) |  | Registry. |
| ingress | [bool](#bool) |  | if true, the cluster has an ingress controller. |






<a name="api-v1-cluster-GetConfigsRequest"></a>

### GetConfigsRequest
Empty Request for getting the configs of all clusters.






<a name="api-v1-cluster-GetConfigsResponse"></a>

### GetConfigsResponse
Empty Response for getting the configs of all clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clusters | [GetConfigResponse](#api-v1-cluster-GetConfigResponse) | repeated |  |






<a name="api-v1-cluster-ListRequest"></a>

### ListRequest
Request for listing available clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






<a name="api-v1-cluster-ListResponse"></a>

### ListResponse
Response for listing available clusters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clusters | [Cluster](#api-v1-cluster-Cluster) | repeated | List of clusters. |






<a name="api-v1-cluster-Registry"></a>

### Registry
Registry dev registry


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | [string](#string) |  |  |








<a name="api-v1-cluster-ClusterType"></a>

### ClusterType
Cluster type - Docker or kubernetes.

| Name | Number | Description |
| ---- | ------ | ----------- |
| CLUSTER_TYPE_UNSPECIFIED | 0 |  |
| CLUSTER_TYPE_DOCKER | 1 |  |
| CLUSTER_TYPE_KUBERNETES | 2 |  |








<a name="api_v1_environment_environment-proto"></a>

## api/v1/environment/environment.proto



<a name="api-v1-environment-Environment"></a>

### Environment
Environment model.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_id | [string](#string) |  | ID of the environment. |
| operator_version | [string](#string) |  | The version of the Rig-operator CRD for this environment. |
| cluster_id | [string](#string) |  | ID of the backing cluster. |
| namespace_template | [string](#string) |  | Namespace template is used to generate the namespace name when configuring resources. Default is to set the namespace equal to the project name. |
| ephemeral | [bool](#bool) |  | If true, the environment is deletable by developer users, and can be deleted with capsules running. |
| active_projects | [string](#string) | repeated | Active Projects. These projects can deploy capsules to this environment. This is overridden by a true the global flag. |
| global | [bool](#bool) |  | If true, the environment is available to all projects. |






<a name="api-v1-environment-Update"></a>

### Update



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| add_project | [string](#string) |  |  |
| remove_project | [string](#string) |  |  |
| set_global | [bool](#bool) |  |  |













<a name="api_v1_environment_revision-proto"></a>

## api/v1/environment/revision.proto



<a name="api-v1-environment-Revision"></a>

### Revision



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| spec | [platform.v1.Environment](#platform-v1-Environment) |  |  |
| metadata | [model.RevisionMetadata](#model-RevisionMetadata) |  |  |













<a name="api_v1_environment_service-proto"></a>

## api/v1/environment/service.proto



<a name="api-v1-environment-CreateRequest"></a>

### CreateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_id | [string](#string) |  |  |
| initializers | [Update](#api-v1-environment-Update) | repeated |  |
| cluster_id | [string](#string) |  |  |
| namespace_template | [string](#string) |  | Namespace template is used to generate the namespace name when configuring resources. Default is to set the namespace equal to the project name. Default value is: {{ .Project.Name }} Valid template properties are: .Project.Name - name of the project .Environment.Name - name of the environment. |
| ephemeral | [bool](#bool) |  | If true, the environment will be marked as ephemeral. It is possible for developers to create ephemeral environments. |






<a name="api-v1-environment-CreateResponse"></a>

### CreateResponse







<a name="api-v1-environment-DeleteRequest"></a>

### DeleteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_id | [string](#string) |  |  |
| force | [bool](#bool) |  | Force delete all running capsules in the enviornment. If false, the request will be aborted if any capsules is running in the environment. |






<a name="api-v1-environment-DeleteResponse"></a>

### DeleteResponse







<a name="api-v1-environment-GetNamespacesRequest"></a>

### GetNamespacesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_envs | [ProjectEnvironment](#api-v1-environment-ProjectEnvironment) | repeated |  |






<a name="api-v1-environment-GetNamespacesResponse"></a>

### GetNamespacesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespaces | [ProjectEnvironmentNamespace](#api-v1-environment-ProjectEnvironmentNamespace) | repeated |  |






<a name="api-v1-environment-GetRequest"></a>

### GetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_id | [string](#string) |  |  |






<a name="api-v1-environment-GetResponse"></a>

### GetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment | [Environment](#api-v1-environment-Environment) |  |  |






<a name="api-v1-environment-ListRequest"></a>

### ListRequest
Request for listing available environments.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |
| exclude_ephemeral | [bool](#bool) |  | Exclude ephemeral environments in the list. |
| project_filter | [string](#string) |  | Get environments for a specific project. |






<a name="api-v1-environment-ListResponse"></a>

### ListResponse
Response for listing available environments.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environments | [Environment](#api-v1-environment-Environment) | repeated | List of environments. |
| platform_version | [string](#string) |  | The version of the Rig-platform. |






<a name="api-v1-environment-ProjectEnvironment"></a>

### ProjectEnvironment



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| environment_id | [string](#string) |  |  |






<a name="api-v1-environment-ProjectEnvironmentNamespace"></a>

### ProjectEnvironmentNamespace



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| environment_id | [string](#string) |  |  |
| namespace | [string](#string) |  |  |






<a name="api-v1-environment-UpdateRequest"></a>

### UpdateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| environment_id | [string](#string) |  |  |
| updates | [Update](#api-v1-environment-Update) | repeated |  |






<a name="api-v1-environment-UpdateResponse"></a>

### UpdateResponse














<a name="model_metadata-proto"></a>

## model/metadata.proto



<a name="model-Metadata"></a>

### Metadata
Generic metadata model.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  | Key of the metadata. |
| value | [bytes](#bytes) |  | Value of the metadata. |













<a name="api_v1_group_group-proto"></a>

## api/v1/group/group.proto



<a name="api-v1-group-Group"></a>

### Group
Group is a named collection of users and service accounts with optional
metadata.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_id | [string](#string) |  | Unique name of the group. |
| num_members | [uint64](#uint64) |  | Number of members. |
| metadata | [Group.MetadataEntry](#api-v1-group-Group-MetadataEntry) | repeated | Metadata of the group. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Creation time of the group. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Last update time of the group. |






<a name="api-v1-group-Group-MetadataEntry"></a>

### Group.MetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bytes](#bytes) |  |  |






<a name="api-v1-group-MemberID"></a>

### MemberID
MemberID is a union of service account id and user id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_account_id | [string](#string) |  | ID of a service account. |
| user_id | [string](#string) |  | ID of a user. |






<a name="api-v1-group-Update"></a>

### Update
An update msg for a group.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_id | [string](#string) |  | Update the unique group name. |
| set_metadata | [model.Metadata](#model-Metadata) |  | Update or create a metadata entry. |
| delete_metadata_key | [string](#string) |  | Delete a metadata entry. |













<a name="api_v1_group_service-proto"></a>

## api/v1/group/service.proto



<a name="api-v1-group-AddMemberRequest"></a>

### AddMemberRequest
Request for adding one or more members to a group.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_id | [string](#string) |  | The group to add members to. |
| member_ids | [MemberID](#api-v1-group-MemberID) | repeated | The members to add. |






<a name="api-v1-group-AddMemberResponse"></a>

### AddMemberResponse
Empty response for adding one or more members to a group.






<a name="api-v1-group-CreateRequest"></a>

### CreateRequest
The request of a Groups.Create RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| initializers | [Update](#api-v1-group-Update) | repeated | The group to create. |






<a name="api-v1-group-CreateResponse"></a>

### CreateResponse
The response of a Groups.Create RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [Group](#api-v1-group-Group) |  | The created group. |






<a name="api-v1-group-DeleteRequest"></a>

### DeleteRequest
The request of a Group.Delete RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_id | [string](#string) |  | The group to delete. |






<a name="api-v1-group-DeleteResponse"></a>

### DeleteResponse
The response of a Group.Delete RPC.






<a name="api-v1-group-GetRequest"></a>

### GetRequest
The request of a Groups.Get RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_id | [string](#string) |  | The group id. |






<a name="api-v1-group-GetResponse"></a>

### GetResponse
The response of a Groups.Get RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [Group](#api-v1-group-Group) |  | The group. |






<a name="api-v1-group-ListGroupsForMemberRequest"></a>

### ListGroupsForMemberRequest
Request for listing the groups a member is in.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| member_id | [MemberID](#api-v1-group-MemberID) |  | The member to list groups for. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






<a name="api-v1-group-ListGroupsForMemberResponse"></a>

### ListGroupsForMemberResponse
Response for listing the groups a member is in.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| groups | [Group](#api-v1-group-Group) | repeated | The groups the member is in. |
| total | [uint64](#uint64) |  | The total amount of groups the member is in. |






<a name="api-v1-group-ListMembersRequest"></a>

### ListMembersRequest
Reqyest for listing members of a group.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_id | [string](#string) |  | The group to list members of. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






<a name="api-v1-group-ListMembersResponse"></a>

### ListMembersResponse
Response for listing members of a group.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| members | [model.MemberEntry](#model-MemberEntry) | repeated | The members in the group. |
| total | [uint64](#uint64) |  | The total amount of members in the group. |






<a name="api-v1-group-ListRequest"></a>

### ListRequest
The request of a Groups.List RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |
| search | [string](#string) |  | Search string. |






<a name="api-v1-group-ListResponse"></a>

### ListResponse
The response of a Groups.List RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| groups | [Group](#api-v1-group-Group) | repeated | list of groups. |
| total | [uint64](#uint64) |  | total amount of groups. |






<a name="api-v1-group-RemoveMemberRequest"></a>

### RemoveMemberRequest
Request for removing a member from a group.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group_id | [string](#string) |  | The group to remove the member from. |
| member_id | [MemberID](#api-v1-group-MemberID) |  | The member to remove. |






<a name="api-v1-group-RemoveMemberResponse"></a>

### RemoveMemberResponse
Empty response for removing a member from a group.






<a name="api-v1-group-UpdateRequest"></a>

### UpdateRequest
The request of a Groups.Update RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| updates | [Update](#api-v1-group-Update) | repeated | The updates to apply. |
| group_id | [string](#string) |  | The group id. |






<a name="api-v1-group-UpdateResponse"></a>

### UpdateResponse
The response of a Groups.Update RPC.













<a name="api_v1_image_service-proto"></a>

## api/v1/image/service.proto



<a name="api-v1-image-AddRequest"></a>

### AddRequest
Request to add a new image in a capsule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule to add the image in. |
| image | [string](#string) |  | Container image to add the image from. |
| digest | [string](#string) |  | Digest of the image. |
| origin | [api.v1.capsule.Origin](#api-v1-capsule-Origin) |  | Origin of the image |
| labels | [AddRequest.LabelsEntry](#api-v1-image-AddRequest-LabelsEntry) | repeated | Meta data to attach to the image. |
| skip_image_check | [bool](#bool) |  | if true skip check if image exists. |
| project_id | [string](#string) |  | Project ID. |






<a name="api-v1-image-AddRequest-LabelsEntry"></a>

### AddRequest.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="api-v1-image-AddResponse"></a>

### AddResponse
Response to add a new image in a capsule.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| image_id | [string](#string) |  | ID of the image. |
| added_new_image | [bool](#bool) |  | True if a new image was added, false if the image already existed. |






<a name="api-v1-image-DeleteRequest"></a>

### DeleteRequest
Request to delete a image.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule to delete the image from. |
| image_id | [string](#string) |  | Image to delete. |
| project_id | [string](#string) |  | Project ID. |






<a name="api-v1-image-DeleteResponse"></a>

### DeleteResponse
Empty response to delete a image.






<a name="api-v1-image-GetImageInfoRequest"></a>

### GetImageInfoRequest
Request to get information about an image.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| image | [string](#string) |  | The image to get information about. |






<a name="api-v1-image-GetImageInfoResponse"></a>

### GetImageInfoResponse
Reponse to GetImageInfo request, containing information about an image.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| image_id | [ImageId](#api-v1-image-ImageId) |  | Image ID. |
| image_string | [string](#string) |  | Image from the request. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the image was created. |
| origin | [api.v1.capsule.Origin](#api-v1-capsule-Origin) |  | Origin of the image. |






<a name="api-v1-image-GetRepositoryInfoRequest"></a>

### GetRepositoryInfoRequest
Get repository information request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registry | [string](#string) |  | Docker Registry |
| repository | [string](#string) |  | Docker Repository |






<a name="api-v1-image-GetRepositoryInfoResponse"></a>

### GetRepositoryInfoResponse
Get repository information response.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tags | [Tag](#api-v1-image-Tag) | repeated | Image Tags in the repository. |






<a name="api-v1-image-GetRequest"></a>

### GetRequest
Request to get a image.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule to get the image from. |
| image_id | [string](#string) |  | Image to get. |
| project_id | [string](#string) |  | Project ID. |






<a name="api-v1-image-GetResponse"></a>

### GetResponse
Response to get a image.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| image | [api.v1.capsule.Image](#api-v1-capsule-Image) |  | The image to retrieve |






<a name="api-v1-image-ImageId"></a>

### ImageId
A collection of image properties that uniquely identifies an image.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registry | [string](#string) |  | Docker Registry. |
| repository | [string](#string) |  | Docker Repository. |
| tag | [string](#string) |  | Tag of the image. |
| digest | [string](#string) |  | Digest of the image. |






<a name="api-v1-image-ListRequest"></a>

### ListRequest
Request to list images.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule to list images in. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |
| project_id | [string](#string) |  | Project ID. |






<a name="api-v1-image-ListResponse"></a>

### ListResponse
Reponse to list images.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| images | [api.v1.capsule.Image](#api-v1-capsule-Image) | repeated | Images in the capsule. |
| total | [uint64](#uint64) |  | Total number of images in the capsule. |






<a name="api-v1-image-Tag"></a>

### Tag
A docker image tag.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tag | [string](#string) |  | Tag of the image. |
| image_created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the image was created. |













<a name="api_v1_metrics_metrics-proto"></a>

## api/v1/metrics/metrics.proto



<a name="api-v1-metrics-Keys"></a>

### Keys



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [bool](#bool) |  |  |
| environment | [bool](#bool) |  |  |
| capsule | [bool](#bool) |  |  |
| metric_name | [bool](#bool) |  |  |
| all | [bool](#bool) |  |  |






<a name="api-v1-metrics-Metric"></a>

### Metric



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| value | [double](#double) |  |  |






<a name="api-v1-metrics-MetricFull"></a>

### MetricFull



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metric | [Metric](#api-v1-metrics-Metric) |  |  |
| tags | [Tags](#api-v1-metrics-Tags) |  |  |






<a name="api-v1-metrics-Tags"></a>

### Tags



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  |  |
| environment | [string](#string) |  |  |
| capsule | [string](#string) |  |  |
| metric_name | [string](#string) |  |  |













<a name="api_v1_metrics_service-proto"></a>

## api/v1/metrics/service.proto



<a name="api-v1-metrics-Aggregation"></a>

### Aggregation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| aggregator | [Aggregator](#api-v1-metrics-Aggregator) |  |  |
| bucket_size | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="api-v1-metrics-Expression"></a>

### Expression



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| leaf | [Expression.Leaf](#api-v1-metrics-Expression-Leaf) |  |  |
| operation | [Expression.Operation](#api-v1-metrics-Expression-Operation) |  |  |
| constant | [Expression.Constant](#api-v1-metrics-Expression-Constant) |  |  |
| with_default | [Expression.WithDefault](#api-v1-metrics-Expression-WithDefault) |  |  |
| sum | [Expression.Sum](#api-v1-metrics-Expression-Sum) |  |  |






<a name="api-v1-metrics-Expression-Constant"></a>

### Expression.Constant



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| constant | [double](#double) |  |  |






<a name="api-v1-metrics-Expression-Leaf"></a>

### Expression.Leaf



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tags | [Tags](#api-v1-metrics-Tags) |  |  |
| aggregator | [Aggregator](#api-v1-metrics-Aggregator) |  |  |






<a name="api-v1-metrics-Expression-Operation"></a>

### Expression.Operation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| left | [Expression](#api-v1-metrics-Expression) |  |  |
| right | [Expression](#api-v1-metrics-Expression) |  |  |
| operation | [BinOp](#api-v1-metrics-BinOp) |  |  |
| on | [Keys](#api-v1-metrics-Keys) |  |  |
| ignore | [Keys](#api-v1-metrics-Keys) |  |  |






<a name="api-v1-metrics-Expression-Sum"></a>

### Expression.Sum



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| on | [Keys](#api-v1-metrics-Keys) |  |  |
| ignore | [Keys](#api-v1-metrics-Keys) |  |  |
| expression | [Expression](#api-v1-metrics-Expression) |  |  |






<a name="api-v1-metrics-Expression-WithDefault"></a>

### Expression.WithDefault



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| expression | [Expression](#api-v1-metrics-Expression) |  |  |
| default | [double](#double) |  |  |






<a name="api-v1-metrics-GetMetricsExpressionRequest"></a>

### GetMetricsExpressionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| expression | [Expression](#api-v1-metrics-Expression) |  |  |
| from | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| to | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| bucket_size | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="api-v1-metrics-GetMetricsExpressionResponse"></a>

### GetMetricsExpressionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metrics | [Metric](#api-v1-metrics-Metric) | repeated |  |






<a name="api-v1-metrics-GetMetricsManyRequest"></a>

### GetMetricsManyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tags | [Tags](#api-v1-metrics-Tags) | repeated |  |
| from | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| to | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| aggregation | [Aggregation](#api-v1-metrics-Aggregation) |  |  |






<a name="api-v1-metrics-GetMetricsManyResponse"></a>

### GetMetricsManyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metrics | [MetricFull](#api-v1-metrics-MetricFull) | repeated |  |






<a name="api-v1-metrics-GetMetricsRequest"></a>

### GetMetricsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tags | [Tags](#api-v1-metrics-Tags) |  |  |
| from | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| to | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| aggregation | [Aggregation](#api-v1-metrics-Aggregation) |  |  |






<a name="api-v1-metrics-GetMetricsResponse"></a>

### GetMetricsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metrics | [Metric](#api-v1-metrics-Metric) | repeated |  |








<a name="api-v1-metrics-Aggregator"></a>

### Aggregator


| Name | Number | Description |
| ---- | ------ | ----------- |
| AGGREGATOR_UNSPECIFIED | 0 |  |
| AGGREGATOR_AVG | 1 |  |
| AGGREGATOR_MIN | 2 |  |
| AGGREGATOR_MAX | 3 |  |
| AGGREGATOR_SUM | 4 |  |



<a name="api-v1-metrics-BinOp"></a>

### BinOp


| Name | Number | Description |
| ---- | ------ | ----------- |
| BINOP_UNSPECIFIED | 0 |  |
| BINOP_ADD | 1 |  |
| BINOP_SUB | 2 |  |
| BINOP_MULT | 3 |  |
| BINOP_DIV | 4 |  |








<a name="model_notification-proto"></a>

## model/notification.proto



<a name="model-NotificationNotifier"></a>

### NotificationNotifier



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| target | [NotificationTarget](#model-NotificationTarget) |  |  |
| topics | [NotificationTopic](#model-NotificationTopic) | repeated |  |
| environments | [EnvironmentFilter](#model-EnvironmentFilter) |  |  |






<a name="model-NotificationTarget"></a>

### NotificationTarget



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| slack | [NotificationTarget.SlackTarget](#model-NotificationTarget-SlackTarget) |  |  |
| email | [NotificationTarget.EmailTarget](#model-NotificationTarget-EmailTarget) |  |  |






<a name="model-NotificationTarget-EmailTarget"></a>

### NotificationTarget.EmailTarget



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| from_email | [string](#string) |  |  |
| to_emails | [string](#string) | repeated |  |






<a name="model-NotificationTarget-SlackTarget"></a>

### NotificationTarget.SlackTarget



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace | [string](#string) |  |  |
| channel_id | [string](#string) |  |  |








<a name="model-NotificationTopic"></a>

### NotificationTopic


| Name | Number | Description |
| ---- | ------ | ----------- |
| NOTIFICATION_TOPIC_UNSPECIFIED | 0 |  |
| NOTIFICATION_TOPIC_ROLLOUT | 1 |  |
| NOTIFICATION_TOPIC_ISSUE | 2 |  |
| NOTIFICATION_TOPIC_PROJECT | 3 |  |
| NOTIFICATION_TOPIC_ENVIRONMENT | 4 |  |
| NOTIFICATION_TOPIC_CAPSULE | 5 |  |
| NOTIFICATION_TOPIC_USER | 6 |  |








<a name="api_v1_project_project-proto"></a>

## api/v1/project/project.proto



<a name="api-v1-project-NotificationNotifiers"></a>

### NotificationNotifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| disabled | [bool](#bool) |  | If the notifiers are disabled, notifiers from parent are not inherited even if these notifiers at this level are empty. |
| notifiers | [model.NotificationNotifier](#model-NotificationNotifier) | repeated |  |






<a name="api-v1-project-Project"></a>

### Project
The top most model that capsules etc belong to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  | The unique id of the project. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the project was created. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the project was last updated. |
| installation_id | [string](#string) |  | The installation id of the project. |
| git_store | [model.GitStore](#model-GitStore) |  |  |
| notifiers | [NotificationNotifiers](#api-v1-project-NotificationNotifiers) |  | The notifiers for the project. |
| pipelines | [model.Pipeline](#model-Pipeline) | repeated | Environment pipelines for the project |






<a name="api-v1-project-Update"></a>

### Update
Update msg for a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| set_git_store | [model.GitStore](#model-GitStore) |  | Set the git store. |
| notifiers | [NotificationNotifiers](#api-v1-project-NotificationNotifiers) |  | Set the notifiers. |
| pipelines | [Update.Pipelines](#api-v1-project-Update-Pipelines) |  | Set the pipelines |






<a name="api-v1-project-Update-Pipelines"></a>

### Update.Pipelines



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipelines | [model.Pipeline](#model-Pipeline) | repeated | The pipelines to update. |













<a name="api_v1_project_revision-proto"></a>

## api/v1/project/revision.proto



<a name="api-v1-project-Revision"></a>

### Revision



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| spec | [platform.v1.Project](#platform-v1-Project) |  |  |
| metadata | [model.RevisionMetadata](#model-RevisionMetadata) |  |  |













<a name="api_v1_project_service-proto"></a>

## api/v1/project/service.proto



<a name="api-v1-project-CreateRequest"></a>

### CreateRequest
The request to create a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| initializers | [Update](#api-v1-project-Update) | repeated | The initializers of the project. |
| project_id | [string](#string) |  | ID of the project to create. |






<a name="api-v1-project-CreateResponse"></a>

### CreateResponse
The response to Create a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [Project](#api-v1-project-Project) |  | The created project. |






<a name="api-v1-project-DeleteRequest"></a>

### DeleteRequest
Request to delete a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  | Id of the project to delete |






<a name="api-v1-project-DeleteResponse"></a>

### DeleteResponse
Empty response for deleting a project.






<a name="api-v1-project-GetCustomObjectMetricsRequest"></a>

### GetCustomObjectMetricsRequest
Request to get custom metrics for a project and environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_reference | [model.ObjectReference](#model-ObjectReference) |  | The object to get metrics for. |
| project_id | [string](#string) |  | The project to get metrics for. |
| environment_id | [string](#string) |  | The environment to get metrics for. |






<a name="api-v1-project-GetCustomObjectMetricsResponse"></a>

### GetCustomObjectMetricsResponse
Response for getting custom metrics for a project and environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metrics | [model.Metric](#model-Metric) | repeated | The metrics for the given object. |
| project_id | [string](#string) |  | The project the metrics are for. |
| environment_id | [string](#string) |  | The environment the metrics are for. |






<a name="api-v1-project-GetEffectiveGitSettingsRequest"></a>

### GetEffectiveGitSettingsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |






<a name="api-v1-project-GetEffectiveGitSettingsResponse"></a>

### GetEffectiveGitSettingsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| git | [model.GitStore](#model-GitStore) |  |  |






<a name="api-v1-project-GetEffectiveNotificationSettingsRequest"></a>

### GetEffectiveNotificationSettingsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |






<a name="api-v1-project-GetEffectiveNotificationSettingsResponse"></a>

### GetEffectiveNotificationSettingsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| notifiers | [model.NotificationNotifier](#model-NotificationNotifier) | repeated |  |






<a name="api-v1-project-GetEffectivePipelineSettingsRequest"></a>

### GetEffectivePipelineSettingsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  |  |
| capsule_id | [string](#string) |  | If set, the response will contain information as to whether the pipeline is already running for that capsule. |






<a name="api-v1-project-GetEffectivePipelineSettingsResponse"></a>

### GetEffectivePipelineSettingsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipelines | [GetEffectivePipelineSettingsResponse.Pipeline](#api-v1-project-GetEffectivePipelineSettingsResponse-Pipeline) | repeated |  |






<a name="api-v1-project-GetEffectivePipelineSettingsResponse-Pipeline"></a>

### GetEffectivePipelineSettingsResponse.Pipeline



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [model.Pipeline](#model-Pipeline) |  |  |
| already_running | [bool](#bool) |  |  |






<a name="api-v1-project-GetObjectsByKindRequest"></a>

### GetObjectsByKindRequest
Request to get all objects of a given kind in a project and environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  | The kind of the objects to get. |
| api_version | [string](#string) |  | The api version of the objects to get. |
| project_id | [string](#string) |  | The project to get the objects for. |
| environment_id | [string](#string) |  | The environment to get the objects for. |






<a name="api-v1-project-GetObjectsByKindResponse"></a>

### GetObjectsByKindResponse
Response for getting all objects of a given kind in a project and
environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| objects | [KubernetesObject](#api-v1-project-KubernetesObject) | repeated | The objects of the given kind. |






<a name="api-v1-project-GetRequest"></a>

### GetRequest
Request for getting a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  | The project to get. |






<a name="api-v1-project-GetResponse"></a>

### GetResponse
Response for getting a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [Project](#api-v1-project-Project) |  | The retrieved project. |






<a name="api-v1-project-KubernetesObject"></a>

### KubernetesObject
Model of a kubernetes object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  | Type / kind of the object. |
| name | [string](#string) |  | Name of the object. |






<a name="api-v1-project-ListRequest"></a>

### ListRequest
Request for listing projects.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






<a name="api-v1-project-ListResponse"></a>

### ListResponse
Response for listing projects.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| projects | [Project](#api-v1-project-Project) | repeated | The retrieved projects. |
| total | [int64](#int64) |  | Total number of projects. |






<a name="api-v1-project-PublicKeyRequest"></a>

### PublicKeyRequest
Request to get the public key of a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  | The project to get the key from. |






<a name="api-v1-project-PublicKeyResponse"></a>

### PublicKeyResponse
Response for getting a projects public key.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| public_key | [string](#string) |  | the retrieved public key. |






<a name="api-v1-project-UpdateRequest"></a>

### UpdateRequest
Update the name field of a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| updates | [Update](#api-v1-project-Update) | repeated | the updates to apply. |
| project_id | [string](#string) |  | The project to update. |






<a name="api-v1-project-UpdateResponse"></a>

### UpdateResponse
Empty response for updating a project.













<a name="api_v1_project_settings_settings-proto"></a>

## api/v1/project/settings/settings.proto



<a name="api-v1-project-settings-AddDockerRegistry"></a>

### AddDockerRegistry
Message for adding a docker registry


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | [string](#string) |  | The host of the docker registry. |
| auth | [string](#string) |  | authentication string to the docker registry. |
| credentials | [DockerRegistryCredentials](#api-v1-project-settings-DockerRegistryCredentials) |  | Credentials for the docker registry. |






<a name="api-v1-project-settings-DockerRegistry"></a>

### DockerRegistry
Docker registry.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secret_id | [string](#string) |  | The secret id of the docker registry. |
| host | [string](#string) |  | Host of the docker registry/ |






<a name="api-v1-project-settings-DockerRegistryCredentials"></a>

### DockerRegistryCredentials
Credentials for a docker registry.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| username | [string](#string) |  | Username for the docker registry. |
| password | [string](#string) |  | Password for the docker registry. |
| email | [string](#string) |  | Email for the docker registry. |






<a name="api-v1-project-settings-Settings"></a>

### Settings
Project wide settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| docker_registries | [DockerRegistry](#api-v1-project-settings-DockerRegistry) | repeated | Docker registries for images. |






<a name="api-v1-project-settings-Update"></a>

### Update
Update message for project settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| add_docker_registry | [AddDockerRegistry](#api-v1-project-settings-AddDockerRegistry) |  | Add a docker registry. |
| delete_docker_registry | [string](#string) |  | Delete a docker registry. |













<a name="api_v1_project_settings_service-proto"></a>

## api/v1/project/settings/service.proto



<a name="api-v1-project-settings-GetLicenseInfoRequest"></a>

### GetLicenseInfoRequest
Request to get the license information of the Rig installation.






<a name="api-v1-project-settings-GetLicenseInfoResponse"></a>

### GetLicenseInfoResponse
Response for getting the license information of the Rig installation.

// The plan of the rig installation.
api.v1.project.Plan plan = 1;
// The expiration date of the license.
google.protobuf.Timestamp expires_at = 2;






<a name="api-v1-project-settings-GetSettingsRequest"></a>

### GetSettingsRequest
Empty get settings request






<a name="api-v1-project-settings-GetSettingsResponse"></a>

### GetSettingsResponse
Response for getting settings for the project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| settings | [Settings](#api-v1-project-settings-Settings) |  | The settings. |






<a name="api-v1-project-settings-UpdateSettingsRequest"></a>

### UpdateSettingsRequest
Request for  updating settings for a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| updates | [Update](#api-v1-project-settings-Update) | repeated | The updates to apply. |






<a name="api-v1-project-settings-UpdateSettingsResponse"></a>

### UpdateSettingsResponse
Empty response for updating a project's settings.













<a name="api_v1_role_role-proto"></a>

## api/v1/role/role.proto



<a name="api-v1-role-EntityID"></a>

### EntityID
EntityID is a oneof type that can be used to represent a user, service
account or group.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  | User entity. |
| service_account_id | [string](#string) |  | Service account entity. |
| group_id | [string](#string) |  | Group entity. |






<a name="api-v1-role-Permission"></a>

### Permission
A permission that is granted to a role.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| action | [string](#string) |  | The action that is action permission to perform. |
| scope | [Scope](#api-v1-role-Scope) |  | The scope in which the action can be performed. |






<a name="api-v1-role-Role"></a>

### Role
Role model for Role based access control.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  | Unique ID of the role. |
| permissions | [Permission](#api-v1-role-Permission) | repeated | The permissions granted to the role. |
| metadata | [Role.MetadataEntry](#api-v1-role-Role-MetadataEntry) | repeated | Metadata associated with the role. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp when the role was created. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp when the role was last updated. |






<a name="api-v1-role-Role-MetadataEntry"></a>

### Role.MetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bytes](#bytes) |  |  |






<a name="api-v1-role-Scope"></a>

### Scope
Scope for permissions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [string](#string) |  | The resource on which the action can be performed. This consists of a type, and an optional ID. fx. "user/*", "group/admin" |
| environment | [string](#string) |  | The environment in which the action can be performed. This can be a wildcard. |
| project | [string](#string) |  | The project in which the action can be performed. This can be a wildcard. |






<a name="api-v1-role-Update"></a>

### Update
Update message to update a field of a role.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| add_permission | [Permission](#api-v1-role-Permission) |  | Adding a permission to the role. |
| remove_permission | [Permission](#api-v1-role-Permission) |  | Removing a permission from the role. |
| set_metadata | [model.Metadata](#model-Metadata) |  | Update or create a metadata field on the role. |
| delete_metadata_key | [string](#string) |  | Delete a metadata field on the role. |













<a name="api_v1_role_service-proto"></a>

## api/v1/role/service.proto



<a name="api-v1-role-AssignRequest"></a>

### AssignRequest
Assign a role to an entity.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  | The role to assign. |
| entity_id | [EntityID](#api-v1-role-EntityID) |  | The entity to assign the role to. |






<a name="api-v1-role-AssignResponse"></a>

### AssignResponse
Empty response of assigning a role to an entity.






<a name="api-v1-role-CreateRequest"></a>

### CreateRequest
Request to create a role.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  | The id / name of the role to create. |
| permissions | [Permission](#api-v1-role-Permission) | repeated | The permissions to assign to the role. |






<a name="api-v1-role-CreateResponse"></a>

### CreateResponse
Response to create a role.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#api-v1-role-Role) |  | The created role. |






<a name="api-v1-role-DeleteRequest"></a>

### DeleteRequest
Request to delete a role.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  | The id / name of the role to delete. |






<a name="api-v1-role-DeleteResponse"></a>

### DeleteResponse
Empty Response to delete a role.






<a name="api-v1-role-GetRequest"></a>

### GetRequest
Request to retrieve a role.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  | The role to retrieve. |






<a name="api-v1-role-GetResponse"></a>

### GetResponse
Response to getting a role.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role | [Role](#api-v1-role-Role) |  | The retrieved role. |






<a name="api-v1-role-ListAssigneesRequest"></a>

### ListAssigneesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  | The role to list assignees for. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






<a name="api-v1-role-ListAssigneesResponse"></a>

### ListAssigneesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entity_ids | [string](#string) | repeated | The assignees of the role. |






<a name="api-v1-role-ListForEntityRequest"></a>

### ListForEntityRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entity_id | [EntityID](#api-v1-role-EntityID) |  | The entity to list roles for. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






<a name="api-v1-role-ListForEntityResponse"></a>

### ListForEntityResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_ids | [string](#string) | repeated | The roles assigned to the entity. |






<a name="api-v1-role-ListRequest"></a>

### ListRequest
Request to list roles.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






<a name="api-v1-role-ListResponse"></a>

### ListResponse
Response to list roles.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| roles | [Role](#api-v1-role-Role) | repeated | The retrieved roles. |






<a name="api-v1-role-RevokeRequest"></a>

### RevokeRequest
Revoke a role from an entity.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  | The role to revoke. |
| entity_id | [EntityID](#api-v1-role-EntityID) |  | The entity to revoke the role from. |






<a name="api-v1-role-RevokeResponse"></a>

### RevokeResponse
Empty response for revoking a role.






<a name="api-v1-role-UpdateRequest"></a>

### UpdateRequest
Request to update a role.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| role_id | [string](#string) |  | the role to update. |
| updates | [Update](#api-v1-role-Update) | repeated | The updates to apply to the role. |






<a name="api-v1-role-UpdateResponse"></a>

### UpdateResponse
Empty update response.













<a name="api_v1_service_account_service_account-proto"></a>

## api/v1/service_account/service_account.proto



<a name="api-v1-service_account-ServiceAccount"></a>

### ServiceAccount
Service account model.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Unique name of the service account. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Creation date. |
| created_by | [model.Author](#model-Author) |  | Author of the service account. |
| managed | [bool](#bool) |  | Whether the service account is managed by the system. |
| group_ids | [string](#string) | repeated | List of group IDs the service account belongs to. |













<a name="api_v1_service_account_service-proto"></a>

## api/v1/service_account/service.proto



<a name="api-v1-service_account-CreateRequest"></a>

### CreateRequest
Request o create a service account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the service account to create. |
| initial_group_id | [string](#string) |  | If set, the service-account will be added to this group upon creation. |






<a name="api-v1-service_account-CreateResponse"></a>

### CreateResponse
Response of creating a service account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_account | [ServiceAccount](#api-v1-service_account-ServiceAccount) |  | The created service account. |
| client_id | [string](#string) |  | The client id of the service account. |
| client_secret | [string](#string) |  | The client secret of the service account. |






<a name="api-v1-service_account-DeleteRequest"></a>

### DeleteRequest
Request to delete a service account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_account_id | [string](#string) |  | The id of the service account to delete. |






<a name="api-v1-service_account-DeleteResponse"></a>

### DeleteResponse
Empty response for deleting a service account.






<a name="api-v1-service_account-ListRequest"></a>

### ListRequest
Request to list service accounts.






<a name="api-v1-service_account-ListResponse"></a>

### ListResponse
Response for listing service accounts.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_accounts | [model.ServiceAccountEntry](#model-ServiceAccountEntry) | repeated | the retrieved service accounts. |













<a name="api_v1_settings_configuration-proto"></a>

## api/v1/settings/configuration.proto



<a name="api-v1-settings-Client"></a>

### Client



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| slack | [Slack](#api-v1-settings-Slack) |  |  |
| email | [EmailClient](#api-v1-settings-EmailClient) | repeated |  |
| git | [Git](#api-v1-settings-Git) | repeated |  |






<a name="api-v1-settings-Configuration"></a>

### Configuration
Platform wide static configuration.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client | [Client](#api-v1-settings-Client) |  |  |
| docker_registries | [string](#string) | repeated |  |
| default_email | [EmailClient](#api-v1-settings-EmailClient) |  |  |
| capsule_extensions | [Configuration.CapsuleExtensionsEntry](#api-v1-settings-Configuration-CapsuleExtensionsEntry) | repeated |  |






<a name="api-v1-settings-Configuration-CapsuleExtensionsEntry"></a>

### Configuration.CapsuleExtensionsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Extension](#api-v1-settings-Extension) |  |  |






<a name="api-v1-settings-EmailClient"></a>

### EmailClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| type | [EmailType](#api-v1-settings-EmailType) |  |  |






<a name="api-v1-settings-Extension"></a>

### Extension



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| json_schema | [string](#string) |  | The schema of the extension, expressed as a json-schema (https://json-schema.org/). |






<a name="api-v1-settings-Git"></a>

### Git



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | [string](#string) |  | URL is a exact match for the repo-url this auth can be used for. |
| url_prefix | [string](#string) |  | URLPrefix is a prefix-match for the repo urls this auth can be used for. |






<a name="api-v1-settings-Slack"></a>

### Slack



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace | [Slack.Workspace](#api-v1-settings-Slack-Workspace) | repeated |  |






<a name="api-v1-settings-Slack-Workspace"></a>

### Slack.Workspace



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |








<a name="api-v1-settings-EmailType"></a>

### EmailType


| Name | Number | Description |
| ---- | ------ | ----------- |
| EMAIL_TYPE_UNSPECIFIED | 0 |  |
| EMAIL_TYPE_MAILJET | 1 |  |
| EMAIL_TYPE_SMTP | 2 |  |








<a name="api_v1_settings_settings-proto"></a>

## api/v1/settings/settings.proto



<a name="api-v1-settings-Settings"></a>

### Settings
Platform wide settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| notification_notifiers | [model.NotificationNotifier](#model-NotificationNotifier) | repeated |  |
| git_store | [model.GitStore](#model-GitStore) |  |  |
| pipelines | [model.Pipeline](#model-Pipeline) | repeated |  |






<a name="api-v1-settings-Update"></a>

### Update
Update message for platform settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| set_notification_notifiers | [Update.SetNotificationNotifiers](#api-v1-settings-Update-SetNotificationNotifiers) |  | Set the notification notifiers. |
| set_git_store | [model.GitStore](#model-GitStore) |  | Set the git store. |
| set_pipelines | [Update.SetPipelines](#api-v1-settings-Update-SetPipelines) |  | Set the pipelines. |






<a name="api-v1-settings-Update-SetNotificationNotifiers"></a>

### Update.SetNotificationNotifiers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| notifiers | [model.NotificationNotifier](#model-NotificationNotifier) | repeated |  |






<a name="api-v1-settings-Update-SetPipelines"></a>

### Update.SetPipelines



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipelines | [model.Pipeline](#model-Pipeline) | repeated |  |








<a name="api-v1-settings-Plan"></a>

### Plan
The plan for a rig installation

| Name | Number | Description |
| ---- | ------ | ----------- |
| PLAN_UNSPECIFIED | 0 | Unspecified / unactivated plan. |
| PLAN_FREE | 1 | Free tier. |
| PLAN_TEAM | 2 | Team / Pro tier. |
| PLAN_ENTERPRISE | 3 | Enterprise tier. |








<a name="model_id-proto"></a>

## model/id.proto



<a name="model-CapsuleID"></a>

### CapsuleID



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  |  |
| environment | [string](#string) |  |  |
| capsule | [string](#string) |  |  |






<a name="model-CapsuleSetID"></a>

### CapsuleSetID



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  |  |
| capsule | [string](#string) |  |  |













<a name="api_v1_settings_service-proto"></a>

## api/v1/settings/service.proto



<a name="api-v1-settings-GetConfigurationRequest"></a>

### GetConfigurationRequest







<a name="api-v1-settings-GetConfigurationResponse"></a>

### GetConfigurationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| configuration | [Configuration](#api-v1-settings-Configuration) |  |  |






<a name="api-v1-settings-GetGitStoreStatusRequest"></a>

### GetGitStoreStatusRequest







<a name="api-v1-settings-GetGitStoreStatusResponse"></a>

### GetGitStoreStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repositories | [GetGitStoreStatusResponse.RepoGitStatus](#api-v1-settings-GetGitStoreStatusResponse-RepoGitStatus) | repeated |  |
| capsules | [GetGitStoreStatusResponse.CapsuleStatus](#api-v1-settings-GetGitStoreStatusResponse-CapsuleStatus) | repeated |  |
| capsule_sets | [GetGitStoreStatusResponse.CapsuleSetStatus](#api-v1-settings-GetGitStoreStatusResponse-CapsuleSetStatus) | repeated |  |
| errors | [GetGitStoreStatusResponse.CallbackErr](#api-v1-settings-GetGitStoreStatusResponse-CallbackErr) | repeated |  |






<a name="api-v1-settings-GetGitStoreStatusResponse-CallbackErr"></a>

### GetGitStoreStatusResponse.CallbackErr



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| err | [string](#string) |  |  |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="api-v1-settings-GetGitStoreStatusResponse-CapsuleSetStatus"></a>

### GetGitStoreStatusResponse.CapsuleSetStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule | [model.CapsuleSetID](#model-CapsuleSetID) |  |  |
| status | [model.GitStatus](#model-GitStatus) |  |  |






<a name="api-v1-settings-GetGitStoreStatusResponse-CapsuleStatus"></a>

### GetGitStoreStatusResponse.CapsuleStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule | [model.CapsuleID](#model-CapsuleID) |  |  |
| status | [model.GitStatus](#model-GitStatus) |  |  |






<a name="api-v1-settings-GetGitStoreStatusResponse-RepoGitStatus"></a>

### GetGitStoreStatusResponse.RepoGitStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repo | [model.RepoBranch](#model-RepoBranch) |  |  |
| status | [model.GitStatus](#model-GitStatus) |  |  |






<a name="api-v1-settings-GetLicenseInfoRequest"></a>

### GetLicenseInfoRequest
Request to get the license information of the Rig installation.






<a name="api-v1-settings-GetLicenseInfoResponse"></a>

### GetLicenseInfoResponse
Response for getting the license information of the Rig installation.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| expires_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The expiration date of the license. |
| user_limit | [int32](#int32) |  | The number of users allowed in the installation. |






<a name="api-v1-settings-GetSettingsRequest"></a>

### GetSettingsRequest







<a name="api-v1-settings-GetSettingsResponse"></a>

### GetSettingsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| settings | [Settings](#api-v1-settings-Settings) |  |  |






<a name="api-v1-settings-UpdateSettingsRequest"></a>

### UpdateSettingsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| updates | [Update](#api-v1-settings-Update) | repeated |  |






<a name="api-v1-settings-UpdateSettingsResponse"></a>

### UpdateSettingsResponse














<a name="api_v1_tunnel_service-proto"></a>

## api/v1/tunnel/service.proto



<a name="api-v1-tunnel-TunnelClose"></a>

### TunnelClose



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tunnel_id | [uint64](#uint64) |  |  |
| code | [uint32](#uint32) |  |  |
| message | [string](#string) |  |  |






<a name="api-v1-tunnel-TunnelData"></a>

### TunnelData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tunnel_id | [uint64](#uint64) |  |  |
| data | [bytes](#bytes) |  |  |






<a name="api-v1-tunnel-TunnelInfo"></a>

### TunnelInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tunnel_id | [uint64](#uint64) |  |  |
| port | [uint32](#uint32) |  |  |






<a name="api-v1-tunnel-TunnelMessage"></a>

### TunnelMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| new_tunnel | [TunnelInfo](#api-v1-tunnel-TunnelInfo) |  |  |
| data | [TunnelData](#api-v1-tunnel-TunnelData) |  |  |
| close | [TunnelClose](#api-v1-tunnel-TunnelClose) |  |  |






<a name="api-v1-tunnel-TunnelRequest"></a>

### TunnelRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [TunnelMessage](#api-v1-tunnel-TunnelMessage) |  |  |






<a name="api-v1-tunnel-TunnelResponse"></a>

### TunnelResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [TunnelMessage](#api-v1-tunnel-TunnelMessage) |  |  |













<a name="api_v1_user_user-proto"></a>

## api/v1/user/user.proto



<a name="api-v1-user-AuthMethod"></a>

### AuthMethod
how a user is authenticated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login_type | [model.LoginType](#model-LoginType) |  | Login type of the user. |






<a name="api-v1-user-Profile"></a>

### Profile
User profile


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| first_name | [string](#string) |  | First name of the user. |
| last_name | [string](#string) |  | Last name of the user. |






<a name="api-v1-user-Session"></a>

### Session
A user's sessions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auth_method | [AuthMethod](#api-v1-user-AuthMethod) |  | how the user is authenticated. |
| is_invalidated | [bool](#bool) |  | if the session is invalidated |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp when the session was created. |
| invalidated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp when the session was invalidated. |
| expires_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp when the session expires. |
| renewed_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp when the session was renewed. |
| country | [string](#string) |  | Country of the session. |
| postal_code | [int32](#int32) |  | Postal code of the session. |






<a name="api-v1-user-SessionEntry"></a>

### SessionEntry
Session entry


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [string](#string) |  | Session ID of the session. |
| session | [Session](#api-v1-user-Session) |  | Session of the session. |






<a name="api-v1-user-Update"></a>

### Update
Update message to update a user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| email | [string](#string) |  | Email of the user. |
| username | [string](#string) |  | Username of the user. |
| phone_number | [string](#string) |  | Deprecated: text is not supported - Phone number of the user. |
| password | [string](#string) |  | Password of the user. |
| profile | [Profile](#api-v1-user-Profile) |  | Profile of the user. |
| is_email_verified | [bool](#bool) |  | Whether the user's email is verified. |
| is_phone_verified | [bool](#bool) |  | Deprecated: text is not supported - Whether the user's phone number is verified. |
| reset_sessions | [Update.ResetSessions](#api-v1-user-Update-ResetSessions) |  | Reset sessions of the user. |
| set_metadata | [model.Metadata](#model-Metadata) |  | Set metadata of the user. |
| delete_metadata_key | [string](#string) |  | Delete metadata of the user. |
| hashed_password | [model.HashingInstance](#model-HashingInstance) |  | Hashed password of the user. |






<a name="api-v1-user-Update-ResetSessions"></a>

### Update.ResetSessions
if sessions are reset, all sessions will be invalidated and a new session
will be created.






<a name="api-v1-user-User"></a>

### User
The user model.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  | User ID of the user. |
| user_info | [model.UserInfo](#model-UserInfo) |  | User info of the user. |
| profile | [Profile](#api-v1-user-Profile) |  | Profile of the user. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp when the user was last updated. |
| register_info | [model.RegisterInfo](#model-RegisterInfo) |  | Register info of the user. |
| is_phone_verified | [bool](#bool) |  | Deprecated: text is not supported - Whether the user's phone number is verified. |
| is_email_verified | [bool](#bool) |  | Whether the user's email is verified. |
| new_sessions_since | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp when the user last created a new session. |
| metadata | [User.MetadataEntry](#api-v1-user-User-MetadataEntry) | repeated | Metadata of the user. |






<a name="api-v1-user-User-MetadataEntry"></a>

### User.MetadataEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bytes](#bytes) |  |  |






<a name="api-v1-user-VerificationCode"></a>

### VerificationCode
short-lived verification code.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| code | [model.HashingInstance](#model-HashingInstance) |  | Hashed verification code. |
| sent_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp when the verification code was sent. |
| expires_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp when the verification code expires. |
| attempts | [int32](#int32) |  | Number of attempts to verify the code. |
| last_attempt | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp of the last attempt to verify the code. |
| type | [VerificationType](#api-v1-user-VerificationType) |  | Type of verification code. |
| user_id | [string](#string) |  | User ID of the user who the code was sent to. |








<a name="api-v1-user-VerificationType"></a>

### VerificationType
Type of verification code

| Name | Number | Description |
| ---- | ------ | ----------- |
| VERIFICATION_TYPE_UNSPECIFIED | 0 | Default value |
| VERIFICATION_TYPE_EMAIL | 1 | Email verification code. |
| VERIFICATION_TYPE_TEXT | 2 | Deprecated: text is not supported - text verification code. |
| VERIFICATION_TYPE_RESET_PASSWORD | 3 | reset password verification code. |








<a name="api_v1_user_service-proto"></a>

## api/v1/user/service.proto



<a name="api-v1-user-CreateRequest"></a>

### CreateRequest
The request to create a user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| initializers | [Update](#api-v1-user-Update) | repeated | Initial fields to set. |
| initial_group_id | [string](#string) |  | If set, the user will be added to this group upon creation. |






<a name="api-v1-user-CreateResponse"></a>

### CreateResponse
The response of creating a user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#api-v1-user-User) |  | The created user. |






<a name="api-v1-user-DeleteRequest"></a>

### DeleteRequest
Request for deleting a user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  | The user identifier to fetch the user. |






<a name="api-v1-user-DeleteResponse"></a>

### DeleteResponse
Empty response for deleting a user.






<a name="api-v1-user-GetByIdentifierRequest"></a>

### GetByIdentifierRequest
Request to get a user by an identifier.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| identifier | [model.UserIdentifier](#model-UserIdentifier) |  | The identifier to lookup. |






<a name="api-v1-user-GetByIdentifierResponse"></a>

### GetByIdentifierResponse
Response to get a user by an identifier.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#api-v1-user-User) |  | The user. |






<a name="api-v1-user-GetRequest"></a>

### GetRequest
Get request for retrieving a user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  | The user identifier to fetch the user. |






<a name="api-v1-user-GetResponse"></a>

### GetResponse
The response of getting a user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#api-v1-user-User) |  | The retrieved user. |






<a name="api-v1-user-ListRequest"></a>

### ListRequest
Request for listing users.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |
| search | [string](#string) |  | Search string. |






<a name="api-v1-user-ListResponse"></a>

### ListResponse
Response for listing users.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [model.UserEntry](#model-UserEntry) | repeated | The users returned. |
| total | [uint64](#uint64) |  | total number of users. |






<a name="api-v1-user-ListSessionsRequest"></a>

### ListSessionsRequest
Request to list a users login sessions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  | The user to retrieve sessions for. |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






<a name="api-v1-user-ListSessionsResponse"></a>

### ListSessionsResponse
The response of listing a users login sessions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sessions | [SessionEntry](#api-v1-user-SessionEntry) | repeated | The retrieved sessions. |
| total | [uint64](#uint64) |  | The total number of sessions. |






<a name="api-v1-user-UpdateRequest"></a>

### UpdateRequest
The request of updating a user.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  | The user identifier of the user to update. |
| updates | [Update](#api-v1-user-Update) | repeated | The updates to apply to the user. |






<a name="api-v1-user-UpdateResponse"></a>

### UpdateResponse
Empty update response.













<a name="model_credentials-proto"></a>

## model/credentials.proto



<a name="model-ProviderCredentials"></a>

### ProviderCredentials
Generic credentials model.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| public_key | [string](#string) |  | Public key. |
| private_key | [string](#string) |  | Private key. |













<a name="api_v1_user_settings_settings-proto"></a>

## api/v1/user/settings/settings.proto



<a name="api-v1-user-settings-AuthMethod"></a>

### AuthMethod
Message that tells how the user was authenticated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login_type | [model.LoginType](#model-LoginType) |  |  |






<a name="api-v1-user-settings-DefaultInstance"></a>

### DefaultInstance
Default email provider instance.






<a name="api-v1-user-settings-EmailInstance"></a>

### EmailInstance
Type of email instance in a provider.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| default | [DefaultInstance](#api-v1-user-settings-DefaultInstance) |  | default from platform config. |
| mailjet | [MailjetInstance](#api-v1-user-settings-MailjetInstance) |  | Mailjet instance. |
| smtp | [SmtpInstance](#api-v1-user-settings-SmtpInstance) |  | SMTP instance. |






<a name="api-v1-user-settings-EmailProvider"></a>

### EmailProvider
The email provider.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | [string](#string) |  | The email-address that the provider sends emails from. |
| credentials | [model.ProviderCredentials](#model-ProviderCredentials) |  | The credentials for the provider. |
| instance | [EmailInstance](#api-v1-user-settings-EmailInstance) |  | The instance of the provider. |






<a name="api-v1-user-settings-EmailProviderEntry"></a>

### EmailProviderEntry
an entry model for the email provider.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | [string](#string) |  | The email-address that the provider sends emails from. |
| client_id | [string](#string) |  | The client id for the provider. |
| secret_id | [string](#string) |  | The secret id for the provider. |
| instance | [EmailInstance](#api-v1-user-settings-EmailInstance) |  | The instance of the provider. |






<a name="api-v1-user-settings-MailjetInstance"></a>

### MailjetInstance
Mailjet email rpvoider instance.






<a name="api-v1-user-settings-RegisterMethod"></a>

### RegisterMethod
Message that tells how the user was registered / created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| system | [RegisterMethod.System](#api-v1-user-settings-RegisterMethod-System) |  | The user was created by the system. |
| signup | [RegisterMethod.Signup](#api-v1-user-settings-RegisterMethod-Signup) |  | The user was self-registered with a login-type. |






<a name="api-v1-user-settings-RegisterMethod-Signup"></a>

### RegisterMethod.Signup
The user was self-registered with a login-type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login_type | [model.LoginType](#model-LoginType) |  | The login type used to register. |






<a name="api-v1-user-settings-RegisterMethod-System"></a>

### RegisterMethod.System
The user was created by the system.






<a name="api-v1-user-settings-Settings"></a>

### Settings
The users settings configuration. Settings of everything that has to do with
users.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| allow_register | [bool](#bool) |  | If true, users can self register. |
| is_verified_email_required | [bool](#bool) |  | If true, users must be verified via email to login. |
| is_verified_phone_required | [bool](#bool) |  | Deprecated: Text is not supported - If true, users must be verified via phone to login. |
| access_token_ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  | Access token Time to Live. |
| refresh_token_ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  | Refresh token Time to Live. |
| verification_code_ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  | Verification code Time to Live. |
| password_hashing | [model.HashingConfig](#model-HashingConfig) |  | The hashing config used to hash passwords. |
| login_mechanisms | [model.LoginType](#model-LoginType) | repeated | The allowed login mechanisms. |
| send_welcome_mail | [bool](#bool) |  | If true, send a welcome email to new users. |
| email_provider | [EmailProviderEntry](#api-v1-user-settings-EmailProviderEntry) |  | The email provider. |
| text_provider | [TextProviderEntry](#api-v1-user-settings-TextProviderEntry) |  | Deprecated: Text is not supported - The text provider. |
| templates | [Templates](#api-v1-user-settings-Templates) |  | The templates used for sending emails and texts. |






<a name="api-v1-user-settings-SmtpInstance"></a>

### SmtpInstance
SMTP email provider instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | [string](#string) |  | Host of the smtp server. |
| port | [int64](#int64) |  | Port of the smtp server. |






<a name="api-v1-user-settings-Template"></a>

### Template
A generic template model for sending emails and texts.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| body | [string](#string) |  | The body of the template. |
| subject | [string](#string) |  | The subject of the template. |
| type | [TemplateType](#api-v1-user-settings-TemplateType) |  | The type of the template. |
| format | [string](#string) | repeated | The format of the template. |






<a name="api-v1-user-settings-Templates"></a>

### Templates



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| welcome_email | [Template](#api-v1-user-settings-Template) |  | The welcome email template. |
| welcome_text | [Template](#api-v1-user-settings-Template) |  | Deprecated: Text is not supported - The welcome text template. |
| reset_password_email | [Template](#api-v1-user-settings-Template) |  | The reset password email template. |
| reset_password_text | [Template](#api-v1-user-settings-Template) |  | Deprecated: Text is not supported - The reset password text template. |
| verify_email | [Template](#api-v1-user-settings-Template) |  | The email verification template. |
| verify_phone_number | [Template](#api-v1-user-settings-Template) |  | Deprecated: Text is not supported - The text verification template. |






<a name="api-v1-user-settings-TextInstance"></a>

### TextInstance
Deprecated: Text is not supported - Type of text instance in a provider.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| default | [DefaultInstance](#api-v1-user-settings-DefaultInstance) |  | default from platform config. |
| twilio | [TwilioInstance](#api-v1-user-settings-TwilioInstance) |  | Twilio instance. |






<a name="api-v1-user-settings-TextProvider"></a>

### TextProvider
Deprecated: Text is not supported - The text provider.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | [string](#string) |  | The phone number that the provider sends texts from. |
| credentials | [model.ProviderCredentials](#model-ProviderCredentials) |  | The credentials for the provider. |
| instance | [TextInstance](#api-v1-user-settings-TextInstance) |  | The instance of the provider. |






<a name="api-v1-user-settings-TextProviderEntry"></a>

### TextProviderEntry
Deprecated: Text is not supported - An entry model for the text provider.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | [string](#string) |  | The phone number that the provider sends texts from. |
| client_id | [string](#string) |  | The client id for the provider. |
| secret_id | [string](#string) |  | The secret id for the provider. |
| instance | [TextInstance](#api-v1-user-settings-TextInstance) |  | The instance of the provider. |






<a name="api-v1-user-settings-TwilioInstance"></a>

### TwilioInstance
Deprecated: Text is not supported - Default text provider instance.






<a name="api-v1-user-settings-Update"></a>

### Update
Update message for updating users settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| allow_register | [bool](#bool) |  | If true, users can self register. |
| is_verified_email_required | [bool](#bool) |  | If true, users must be verified via email to login. |
| is_verified_phone_required | [bool](#bool) |  | If true, users must be verified via phone to login. |
| access_token_ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  | Access token Time to Live. |
| refresh_token_ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  | Refresh token Time to Live. |
| verification_code_ttl | [google.protobuf.Duration](#google-protobuf-Duration) |  | Verification code Time to Live. |
| password_hashing | [model.HashingConfig](#model-HashingConfig) |  | The hashing config used to hash passwords. |
| login_mechanisms | [Update.LoginMechanisms](#api-v1-user-settings-Update-LoginMechanisms) |  | The allowed login mechanisms. |
| email_provider | [EmailProvider](#api-v1-user-settings-EmailProvider) |  | The email provider. |
| text_provider | [TextProvider](#api-v1-user-settings-TextProvider) |  | The text provider. |
| template | [Template](#api-v1-user-settings-Template) |  | The templates used for sending emails and texts. |






<a name="api-v1-user-settings-Update-LoginMechanisms"></a>

### Update.LoginMechanisms
The allowed login mechanisms


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login_mechanisms | [model.LoginType](#model-LoginType) | repeated |  |








<a name="api-v1-user-settings-TemplateType"></a>

### TemplateType
The different template types.

| Name | Number | Description |
| ---- | ------ | ----------- |
| TEMPLATE_TYPE_UNSPECIFIED | 0 | Unspecified template type. |
| TEMPLATE_TYPE_WELCOME_EMAIL | 1 | The welcome email template. |
| TEMPLATE_TYPE_EMAIL_VERIFICATION | 2 | The email verification template. |
| TEMPLATE_TYPE_EMAIL_RESET_PASSWORD | 3 | The reset password email template. |
| TEMPLATE_TYPE_WELCOME_TEXT | 4 | Deprecated: Text is not supported - The welcome text template. |
| TEMPLATE_TYPE_TEXT_VERIFICATION | 5 | Deprecated: Text is not supported - The text verification template. |
| TEMPLATE_TYPE_TEXT_RESET_PASSWORD | 6 | Deprecated: Text is not supported - The reset password text template. |








<a name="api_v1_user_settings_service-proto"></a>

## api/v1/user/settings/service.proto



<a name="api-v1-user-settings-GetSettingsRequest"></a>

### GetSettingsRequest
Request for getting users settings for the Rig project.






<a name="api-v1-user-settings-GetSettingsResponse"></a>

### GetSettingsResponse
Response for getting users settings for the Rig project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| settings | [Settings](#api-v1-user-settings-Settings) |  | The users settings. |






<a name="api-v1-user-settings-UpdateSettingsRequest"></a>

### UpdateSettingsRequest
Request for updating users settings for the Rig project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| settings | [Update](#api-v1-user-settings-Update) | repeated | The updates to apply to the users settings. |






<a name="api-v1-user-settings-UpdateSettingsResponse"></a>

### UpdateSettingsResponse
Empty response for updating users settings for the Rig project.













<a name="k8s-io_apimachinery_pkg_api_resource_generated-proto"></a>

## k8s.io/apimachinery/pkg/api/resource/generated.proto



<a name="k8s-io-apimachinery-pkg-api-resource-Quantity"></a>

### Quantity



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| string | [string](#string) |  |  |













<a name="k8s-io_apimachinery_pkg_apis_meta_v1_generated-proto"></a>

## k8s.io/apimachinery/pkg/apis/meta/v1/generated.proto



<a name="k8s-io-apimachinery-pkg-apis-meta-v1-TypeMeta"></a>

### TypeMeta



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  |  |
| aPIVersion | [string](#string) |  |  |












