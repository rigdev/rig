<a name="top"></a>







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
| /api.v1.capsule.Service/ListInstances | [ListInstancesRequest](#api-v1-capsule-ListInstancesRequest) | [ListInstancesResponse](#api-v1-capsule-ListInstancesResponse) | Lists all instances for the capsule. |
| /api.v1.capsule.Service/RestartInstance | [RestartInstanceRequest](#api-v1-capsule-RestartInstanceRequest) | [RestartInstanceResponse](#api-v1-capsule-RestartInstanceResponse) | Restart a single capsule instance. |
| /api.v1.capsule.Service/GetRollout | [GetRolloutRequest](#api-v1-capsule-GetRolloutRequest) | [GetRolloutResponse](#api-v1-capsule-GetRolloutResponse) | Get a single rollout by ID. |
| /api.v1.capsule.Service/ListRollouts | [ListRolloutsRequest](#api-v1-capsule-ListRolloutsRequest) | [ListRolloutsResponse](#api-v1-capsule-ListRolloutsResponse) | Lists all rollouts for the capsule. |
| /api.v1.capsule.Service/AbortRollout | [AbortRolloutRequest](#api-v1-capsule-AbortRolloutRequest) | [AbortRolloutResponse](#api-v1-capsule-AbortRolloutResponse) | Abort the rollout. |
| /api.v1.capsule.Service/StopRollout | [StopRolloutRequest](#api-v1-capsule-StopRolloutRequest) | [StopRolloutResponse](#api-v1-capsule-StopRolloutResponse) | Stop a Rollout, removing all resources associated with it. |
| /api.v1.capsule.Service/ListEvents | [ListEventsRequest](#api-v1-capsule-ListEventsRequest) | [ListEventsResponse](#api-v1-capsule-ListEventsResponse) | List capsule events. |
| /api.v1.capsule.Service/CapsuleMetrics | [CapsuleMetricsRequest](#api-v1-capsule-CapsuleMetricsRequest) | [CapsuleMetricsResponse](#api-v1-capsule-CapsuleMetricsResponse) | Get metrics for a capsule |
| /api.v1.capsule.Service/GetInstanceStatus | [GetInstanceStatusRequest](#api-v1-capsule-GetInstanceStatusRequest) | [GetInstanceStatusResponse](#api-v1-capsule-GetInstanceStatusResponse) | GetInstanceStatus returns the current status for the given instance. |
| /api.v1.capsule.Service/ListInstanceStatuses | [ListInstanceStatusesRequest](#api-v1-capsule-ListInstanceStatusesRequest) | [ListInstanceStatusesResponse](#api-v1-capsule-ListInstanceStatusesResponse) | ListInstanceStatuses lists the status of all instances. |
| /api.v1.capsule.Service/Execute | [ExecuteRequest](#api-v1-capsule-ExecuteRequest) stream | [ExecuteResponse](#api-v1-capsule-ExecuteResponse) stream | Execute executes a command in a given in instance, and returns the output along with an exit code. |
| /api.v1.capsule.Service/GetCustomInstanceMetrics | [GetCustomInstanceMetricsRequest](#api-v1-capsule-GetCustomInstanceMetricsRequest) | [GetCustomInstanceMetricsResponse](#api-v1-capsule-GetCustomInstanceMetricsResponse) |  |
| /api.v1.capsule.Service/GetJobExecutions | [GetJobExecutionsRequest](#api-v1-capsule-GetJobExecutionsRequest) | [GetJobExecutionsResponse](#api-v1-capsule-GetJobExecutionsResponse) | Get list of job executions performed by the Capsule. |








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







### api.v1.service_account.Service
<a name="api-v1-service_account-Service"></a>



| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| /api.v1.service_account.Service/Create | [CreateRequest](#api-v1-service_account-CreateRequest) | [CreateResponse](#api-v1-service_account-CreateResponse) | Create a new Service Account. The returned client_id and client_secret can be used as login credentials. Note that the client_secret can only be read out once, at creation. |
| /api.v1.service_account.Service/List | [ListRequest](#api-v1-service_account-ListRequest) | [ListResponse](#api-v1-service_account-ListResponse) | List all service accounts. |
| /api.v1.service_account.Service/Delete | [DeleteRequest](#api-v1-service_account-DeleteRequest) | [DeleteResponse](#api-v1-service_account-DeleteResponse) | Delete a service account. It can take up to the TTL of access tokens for existing sessions using this service_account, to expire. |







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








<a name="model_auth-proto"></a>

## model/auth.proto





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






<a name="api-v1-capsule-Update"></a>

### Update
Legacy update message













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
| object_reference | [ObjectReference](#api-v1-capsule-ObjectReference) |  | Reference to the object. |






<a name="api-v1-capsule-ObjectMetric-MatchLabelsEntry"></a>

### ObjectMetric.MatchLabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="api-v1-capsule-ObjectReference"></a>

### ObjectReference
A reference to a kubernetes object.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  | Type of object. |
| name | [string](#string) |  | Name of the object. |
| api_version | [string](#string) |  | Api version of the object. |






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
| type | [ContainerType](#api-v1-capsule-instance-ContainerType) |  |  |






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








<a name="api-v1-capsule-instance-ContainerType"></a>

### ContainerType


| Name | Number | Description |
| ---- | ------ | ----------- |
| CONTAINER_TYPE_UNSPECIFIED | 0 |  |
| CONTAINER_TYPE_MAIN | 1 |  |
| CONTAINER_TYPE_SIDECAR | 2 |  |
| CONTAINER_TYPE_INIT | 3 |  |



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














<a name="api_v1_capsule_metrics-proto"></a>

## api/v1/capsule/metrics.proto



<a name="api-v1-capsule-ContainerMetrics"></a>

### ContainerMetrics
Metrics for a container.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp of the metrics. |
| memory_bytes | [uint64](#uint64) |  | Memory usage in bytes. |
| cpu_ms | [uint64](#uint64) |  | CPU usage in milliseconds. |
| storage_bytes | [uint64](#uint64) |  | Storage usage in bytes. |






<a name="api-v1-capsule-InstanceMetrics"></a>

### InstanceMetrics
Metrics for an instance


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| capsule_id | [string](#string) |  | Capsule of the instance. |
| instance_id | [string](#string) |  | Instance ID. |
| main_container | [ContainerMetrics](#api-v1-capsule-ContainerMetrics) |  | Main container metrics. |
| proxy_container | [ContainerMetrics](#api-v1-capsule-ContainerMetrics) |  | Proxy container metrics. |













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








<a name="api_v1_capsule_rollout-proto"></a>

## api/v1/capsule/rollout.proto



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






<a name="api-v1-capsule-RolloutConfig"></a>

### RolloutConfig



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created_by | [model.Author](#model-Author) |  | The user who initiated the rollout. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| changes | [Change](#api-v1-capsule-Change) | repeated |  |
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













<a name="api_v1_capsule_service-proto"></a>

## api/v1/capsule/service.proto



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
| instance_metrics | [InstanceMetrics](#api-v1-capsule-InstanceMetrics) | repeated | Metrics |






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
| force_override | [bool](#bool) |  | By default, existing objects will be kept in favor of overriding them. To force the override of resources, set this flag to true. An example of this use-case is a migration step, where resource created by a previous toolchain e.g. based on Helm charts, are to be replaced and instead be created by the Rig operator. While the override is irreversible, this flag is not "sticky" and must be set by each deploy that should use this behavior. |






<a name="api-v1-capsule-DeployResponse"></a>

### DeployResponse
Deploy response.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rollout_id | [uint64](#uint64) |  | ID of the new rollout. |
| resource_yaml | [DeployResponse.ResourceYamlEntry](#api-v1-capsule-DeployResponse-ResourceYamlEntry) | repeated | The YAML of the resources that will be deployed. |
| rollout_config | [RolloutConfig](#api-v1-capsule-RolloutConfig) |  | The rollout config. |






<a name="api-v1-capsule-DeployResponse-ResourceYamlEntry"></a>

### DeployResponse.ResourceYamlEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






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
| metrics | [Metric](#api-v1-capsule-Metric) | repeated | Custom Metrics. |






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






<a name="api-v1-capsule-Metric"></a>

### Metric
Custom metrics


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the metric. |
| latest_value | [double](#double) |  | Latest value of the metric. |
| latest_timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Timestamp of the latest value. |






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
| default | [bool](#bool) |  | If true, this is the default environment. |
| operator_version | [string](#string) |  | The version of the Rig-operator CRD for this environment. |
| cluster_id | [string](#string) |  | ID of the backing cluster. |
| namespace_template | [string](#string) |  | Namespace template is used to generate the namespace name when configuring resources. Default is to set the namespace equal to the project name. |






<a name="api-v1-environment-Update"></a>

### Update



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| default | [bool](#bool) |  |  |













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






<a name="api-v1-environment-ListRequest"></a>

### ListRequest
Request for listing available environments.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [model.Pagination](#model-Pagination) |  | Pagination options. |






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













<a name="api_v1_project_project-proto"></a>

## api/v1/project/project.proto



<a name="api-v1-project-Project"></a>

### Project
The top most model that capsules etc belong to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  | The unique id of the project. |
| name | [string](#string) |  | Deprecated: Name of the project. |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the project was created. |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | When the project was last updated. |
| installation_id | [string](#string) |  | The installation id of the project. |






<a name="api-v1-project-Update"></a>

### Update
Update msg for a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Update the name of the project. |








<a name="api-v1-project-Plan"></a>

### Plan
The plan for a rig installation

| Name | Number | Description |
| ---- | ------ | ----------- |
| PLAN_UNSPECIFIED | 0 | Unspecified / unactivated plan. |
| PLAN_FREE | 1 | Free tier. |
| PLAN_TEAM | 2 | Team / Pro tier. |
| PLAN_ENTERPRISE | 3 | Enterprise tier. |








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
| object_reference | [api.v1.capsule.ObjectReference](#api-v1-capsule-ObjectReference) |  | The object to get metrics for. |
| project_id | [string](#string) |  | The project to get metrics for. |
| environment_id | [string](#string) |  | The environment to get metrics for. |






<a name="api-v1-project-GetCustomObjectMetricsResponse"></a>

### GetCustomObjectMetricsResponse
Response for getting custom metrics for a project and environment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| metrics | [api.v1.capsule.Metric](#api-v1-capsule-Metric) | repeated | The metrics for the given object. |
| project_id | [string](#string) |  | The project the metrics are for. |
| environment_id | [string](#string) |  | The environment the metrics are for. |






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


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| plan | [api.v1.project.Plan](#api-v1-project-Plan) |  | The plan of the rig installation. |
| expires_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The expiration date of the license. |






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












