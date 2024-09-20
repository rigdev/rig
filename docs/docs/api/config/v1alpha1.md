---
custom_edit_url: null
---


# config.rig.dev/v1alpha1

Package v1alpha1 contains API Schema definitions for the config v1alpha1 API group

## Resource Types
- [OperatorConfig](#operatorconfig)
- [PlatformConfig](#platformconfig)



### Auth



Auth specifies authentication configuration.

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `secret` _string_ | Secret specifies a secret which will be used for jwt signatures. |
| `certificateFile` _string_ | CertificateFile specifies a path to a PEM encoded certificate file which<br />will be used for validating jwt signatures. |
| `certificateKeyFile` _string_ | CertificateKeyFile specifies a path to a PEM encoded certificate key<br />which will be used for jwt signatures. |
| `disablePasswords` _boolean_ | DisablePasswords disables password authentication. This makes sense if<br />you want to require SSO, as login method. |
| `sso` _[SSO](#sso)_ | SSO specifies single sign on configuration. |
| `allowRegister` _boolean_ | AllowRegister specifies if users are allowed to register new accounts. |
| `requireVerification` _boolean_ | IsVerified specifies if users are required to verify their email address. |
| `sendWelcomeEmail` _boolean_ | SendWelcomeEmail specifies if a welcome email should be sent to new users.<br />This will use the default email config |


### CapsuleMatch





_Appears in:_
- [Step](#step)

| Field | Description |
| --- | --- |
| `namespaces` _string array_ | If set, only capsules in one of the namespaces given will have this step run. |
| `names` _string array_ | If set, only execute the plugin on the capsules specified. |
| `annotations` _object (keys:string, values:string)_ | If set, only execute the plugin on the capsules matching the annotations. |
| `enableForPlatform` _boolean_ | If set, will enable the step for the Rig platform which is a Capsule as well |


### CapsuleStep





_Appears in:_
- [Pipeline](#pipeline)

| Field | Description |
| --- | --- |
| `plugin` _string_ | The plugin to use for handling the capsule step.<br />fx. "rigdev.ingress_routes" for routesStep will create an ingress resource per route.<br />fx. "rigdev.deployment" for deploymentStep will use the default deployment plugin. |
| `config` _string_ | Config is a string defining the plugin-specific configuration of the plugin. |


### Client



Client holds various client configuration

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `postgres` _[ClientPostgres](#clientpostgres)_ | Postgres holds configuration for the postgres client. |
| `docker` _[ClientDocker](#clientdocker)_ | Docker sets the host for the Docker client. |
| `mailjet` _[ClientMailjet](#clientmailjet)_ | Deprecated: use 'client.mailjets' instead.<br />Mailjet sets the API key and secret for the Mailjet client. |
| `mailjets` _object (keys:string, values:[ClientMailjet](#clientmailjet))_ | Mailjets holds configuration for multiple mailjet clients.<br />The key is the id of the client, which should be unique across Mailjet and SMTP clients. |
| `smtp` _[ClientSMTP](#clientsmtp)_ | Deprecated: use 'client.smtps' instead.<br />SMTP sets the host, port, username and password for the SMTP client. |
| `smtps` _object (keys:string, values:[ClientSMTP](#clientsmtp))_ | SMTPs holds configuration for muliple SMTP clients.<br />The key is the id of the client, which should be unique across Mailjet and SMTP clients. |
| `operator` _[ClientOperator](#clientoperator)_ | Operator sets the base url for the Operator client. |
| `slack` _object (keys:string, values:[ClientSlack](#clientslack))_ | Slack holds configuration for sending slack messages. The key is the id of the client.<br />For example the workspace in which the app is installed |
| `git` _[ClientGit](#clientgit)_ | Git client configuration for communicating with multiple repositories. |


### ClientDocker



ClientDocker specifies the configuration for the docker client.

_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `host` _string_ | Host where the docker daemon can be reached. |


### ClientGit



ClientGit contains configuration for git integrations.
A given git repository can have authentication from either Auths or GitHubAuths with preference
for GitHubAuths if there is a match.

_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `auths` _[GitAuth](#gitauth) array_ | Auths the git client can behave as. |
| `gitHubAuths` _[GitHub](#github) array_ | GitHubAuths is authentication information for GitHub repositories. |
| `gitLabAuths` _[GitLab](#gitlab) array_ | GitLabAuths is the authentication information for GitLab repositories. |
| `author` _[GitAuthor](#gitauthor)_ | Author used when creating commits. |


### ClientMailjet



ClientMailjet specifes the configuration for the mailjet client.

_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `apiKey` _string_ | APIKey is the mailjet API key |
| `secretKey` _string_ | SecretKey is the mailjet secret key |


### ClientOperator



ClientOperator specifies the configuration for the operator client.

_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `baseUrl` _string_ | BaseURL is the URL used to connect to the operator API |


### ClientPostgres



ClientPostgres specifies the configuration for the postgres client.

_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `user` _string_ | User is the database user used when connecting to the postgres database. |
| `password` _string_ | Password is the password used when connecting to the postgres database. |
| `host` _string_ | Host is the host where the postgres database can be reached. |
| `port` _integer_ | Port is the port of the postgres database server. |
| `database` _string_ | Database in the postgres server to use |
| `insecure` _boolean_ | Insecure is wether to use SSL when connecting to the postgres server |


### ClientSMTP



ClientSMTP specifies the configuration for the SMTP client.

_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `host` _string_ | Host is the SMTP server host. |
| `port` _integer_ | Port is the SMTP server port to use. |
| `username` _string_ | Username used when connecting to the SMTP server. |
| `password` _string_ | Password used when connecting to the SMTP server. |


### ClientSlack





_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `token` _string_ | Slack authentication token. |


### Cluster



Cluster specifies cluster configuration

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `url` _string_ | URL to communicate to the cluster. If set, a Token and CertificateAuthority should<br />be provided as well.<br />If not set, the cluster is interpreted to be the cluster in which the platform runs. |
| `token` _string_ | Token for communicating with the cluster. Available through a service-account's secret. |
| `script` _string_ | Script to execute for getting an access-token to the cluster.<br />The output is expected to be a json-encoding of an `ExecCredential`.<br />See https://pkg.go.dev/k8s.io/client-go@v0.31.0/pkg/apis/clientauthentication/v1beta1#ExecCredential<br />for the format of the struct. |
| `certificateAuthority` _string_ | Certificate authority for communicating with the cluster. Available through a service-account's secret. |
| `type` _[ClusterType](#clustertype)_ | Type of the cluster - either `docker` or `k8s`. |
| `devRegistry` _[DevRegistry](#devregistry)_ | DevRegistry configuration |
| `git` _[ClusterGit](#clustergit)_ | Git sets up gitops write back for this cluster. |
| `createPullSecrets` _boolean_ | If set, secrets will be created if needed, for pulling images. |


### ClusterGit



ClusterGit specifies configuration for git integration. This can be used to
tie rig into a gitops setup.

_Appears in:_
- [Cluster](#cluster)

| Field | Description |
| --- | --- |
| `url` _string_ | URL is the git repository URL. |
| `branch` _string_ | Branch to commit changes to. |
| `pathPrefix` _string_ | PathPrefix path to commit to in git repository.<br />Deprecated: Use `pathPrefixes` instead. |
| `pathPrefixes` _[PathPrefixes](#pathprefixes)_ | PathPrefixes path to commit to in git repository |
| `templates` _[GitTemplates](#gittemplates)_ | Templates used for commit messages. |
| `credentials` _[GitCredentials](#gitcredentials)_ | Credentials to use when connecting to git.<br />Deprecated: Use `client.git.auths` instead. |
| `author` _[GitAuthor](#gitauthor)_ | Author used when creating commits.<br />Deprecated: Use `client.git.author` instead. |


### ClusterType

_Underlying type:_ _string_

ClusterType is a cluster type.

_Appears in:_
- [Cluster](#cluster)



### CustomPlugin





_Appears in:_
- [Pipeline](#pipeline)

| Field | Description |
| --- | --- |
| `image` _string_ | The container image which supplies the plugins |


### DevRegistry



DevRegistry specifies configuration for the dev registry support.

_Appears in:_
- [Cluster](#cluster)

| Field | Description |
| --- | --- |
| `host` _string_ | Host is the host used in image names when pushing to the registry from<br />outside of the cluster. |
| `clusterHost` _string_ | ClusterHost is the host where the registry can be reached from within<br />the cluster. Any image which is named after `Host` will be rename to use<br />`ClusterHost` instead. This ensures that the image can be pulled from<br />within the cluster. |


### DockerRegistryCredentials





_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `username` _string_ | Username for the docker registry. |
| `password` _string_ | Password for the docker registry. |
| `script` _string_ | Script (shell) to execute that should echo the credentials.<br />The output is expected to be a single line (with new-line termination) of format `<username>:<password>`. |
| `expire` _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#duration-v1-meta)_ | Expire is the maximum duration a credential will be cached for, before it's recycled.<br />If a cached credential is rejected before this time, it may be renewed before this duration is expired.<br />Default is `12h`. |


### Email



Email holds configuration for sending emails. Either using mailjet or using SMTP

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `id` _string_ | ID is the specified id an email configuration. |
| `from` _string_ | From is who is set as the sender of rig emails. |
| `type` _[EmailType](#emailtype)_ | Deprecated: ID for an email configuration is used instead. |


### EmailType

_Underlying type:_ _string_

EmailType represents a type of mailing provider

_Appears in:_
- [Email](#email)



### Extension



Extension is a typed (through JSON Schema) expansion of a Platform resource,
that allows extending the default customization.

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `schema` _[JSONSchemaProps](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#jsonschemaprops-v1-apiextensions-k8s-io)_ | The schema of the extension, expressed as a json-schema<br />(https://json-schema.org/). While the full syntax is supported,<br />some features may be semantically disabled which would make the Platform<br />not start or not process Rollouts. |


### GitAuth





_Appears in:_
- [ClientGit](#clientgit)

| Field | Description |
| --- | --- |
| `url` _string_ | URL is a exact match for the repo-url this auth can be used for. |
| `urlPrefix` _string_ | URLPrefix is a prefix-match for the repo urls this auth can be used for.<br />Deprecated; use Match instead |
| `match` _[URLMatch](#urlmatch)_ | How the url should be matched. Can either be 'exact' or 'prefix'<br />Defaults to 'exact' |
| `credentials` _[GitCredentials](#gitcredentials)_ | Credentials to use when connecting to git. |
| `pullingIntervalSeconds` _integer_ | If no web hook is confugured, pull the git repository at the set interval instead<br />to fetch changes. Defaults to 3 mins if no value. |


### GitAuthor



GitAuthor specifies a git commit author

_Appears in:_
- [ClientGit](#clientgit)
- [ClusterGit](#clustergit)

| Field | Description |
| --- | --- |
| `name` _string_ | Name of author |
| `email` _string_ | Email of author |


### GitCredentials



GitCredentials specifies how to authenticate against git.

_Appears in:_
- [ClusterGit](#clustergit)
- [GitAuth](#gitauth)

| Field | Description |
| --- | --- |
| `https` _[HTTPSCredential](#httpscredential)_ | HTTPS specifies basic auth credentials. |
| `ssh` _[SSHCredential](#sshcredential)_ | SSH specifies SSH credentials. |


### GitHub



GitHub contains configuration specifically for GitHub repositories.
To enable pull requests on a GitHub repository, you must add GitHub authentication
using appID, installationID and privateKey for a GitHub app with read/write access to
pull requests.
To have normal read/write access to a repository, you can forego GitHub app authentication
if there is a GitAuth section with credentials for the given repository instead.
If you have GitHub app authentication for a GitHub app with read/write access to the repository,
you don't need a matching GitAuth section.

_Appears in:_
- [ClientGit](#clientgit)

| Field | Description |
| --- | --- |
| `orgRepo` _string_ | OrgRepo is a string containing the GitHub organization and optionally a repository as well.<br />If both org and repo is given, they should be seperated by a '/', e.g. 'myorg/myrepo'.<br />If repo is not given, e.g. 'myrepo', then it matches all repositories within the org 'myorg'.<br />If both org and repo is given, it matches exactly the repo within the org. |
| `organization` _string_ | Organization is the GitHub organization to match.<br />Deprecated. Use OrgRepo instead |
| `repository` _string_ | Repository matches the GitHub repository. If empty, matches all.<br />Deprecated. Use OrgRepo instead |
| `auth` _[GitHubAuth](#githubauth)_ | Auth contains GitHub specific authentication configuration. |
| `polling` _[GitHubPolling](#githubpolling)_ | Polling contains GitHub specific configuration. |


### GitHubAuth



GitHubAuth contains authentication information specifically for a GitHub repository.
Authentication is done using GitHub apps. See https://docs.rig.dev/operator-manual/gitops#github-authentication
for a guide on how to set it up.

_Appears in:_
- [GitHub](#github)

| Field | Description |
| --- | --- |
| `appID` _integer_ | AppID is the app ID of the GitHub app |
| `installationID` _integer_ | InstallationID is the installation ID of the GitHub app |
| `privateKey` _string_ | PrivateKey is a PEM encoded SSH private key. |
| `privateKeyPassword` _string_ | PrivateKeyPassword is an optional password for the SSH private key. |


### GitHubPolling



GitHubPolling defines webhook/pulling configuration for a GitHub repository.

_Appears in:_
- [GitHub](#github)

| Field | Description |
| --- | --- |
| `webhookSecret` _string_ | WebHookSecret is the secret used to validate incoming webhooks. |
| `pullingIntervalSeconds` _integer_ | If webHookSecret isn't set, pull the git repository at the set interval instead<br />to fetch changes. Defaults to 3 mins if no value. |


### GitLab



GitLab contains configuration specifically for GitLab repositories.
To enable pull requests on a GitLab repository, you must add GitLab authentication
using an access token.
To have normal read/write access to a repository, you can forego GitLab access tokens
if there is a GitAuth section with credentials for the given repository instead.
If you have GitLab authentication for a repository, you don't need a matching GitAuth section.

_Appears in:_
- [ClientGit](#clientgit)

| Field | Description |
| --- | --- |
| `groupsProject` _string_ | GroupsProject is a string containing a list of GitLab groups and optionally a project<br />Groups are separated by '/' and project by ':', e.g.<br />group/subgroup1/subgroup2:project<br />If a project is given, it matches exactly that project within that sequence of subsgroups<br />If no project is given, it matches all projects within all subgroups which are children of the<br />given group sequence. E.g.<br />'group' will match 'group/subgroup1:project1' and 'group/subgroup1/subgroup2:project2' |
| `groups` _string array_ | Groups is a sequence of GitLab groups.<br />The first is the main group and the rest a nesting of subgroups.<br />If Project is empty, the configuration will match any<br />GitLab repository whose (group, subgroups) sequence where 'groups' is a prefix.<br />Deprecated. Use GroupsProject |
| `project` _string_ | Project is the GitLab project of the repository. Can be empty for matching all project names.<br />Deprecated. Use GroupsProject |
| `auth` _[GitLabAuth](#gitlabauth)_ | Auth contains GitLab specific authentication configuration. |
| `polling` _[GitLabPolling](#gitlabpolling)_ | Polling contains GitLab specific configuration. |


### GitLabAuth



GitLabAuth contains authentication information specifically for a GitLab repository.
Authentication is done using an access token. See https://docs.rig.dev/operator-manual/gitops#gitlab-authentication
for a guide on how to set it up.

_Appears in:_
- [GitLab](#gitlab)

| Field | Description |
| --- | --- |
| `accessToken` _string_ | AccessToken is an accessToken which is used to authenticate against the GitLab repository. |


### GitLabPolling



GitLabPolling defines webhook/pulling configuration for a GitLab repository.

_Appears in:_
- [GitLab](#gitlab)

| Field | Description |
| --- | --- |
| `webhookSecret` _string_ | WebHookSecret is the secret used to validate incoming webhooks. |
| `pullingIntervalSeconds` _integer_ | If webHookSecret isn't set, pull the git repository at the set interval instead<br />to fetch changes. Defaults to 3 mins if no value. |


### GitTemplates



GitTemplates specifies the templates used for creating commits.

_Appears in:_
- [ClusterGit](#clustergit)

| Field | Description |
| --- | --- |
| `rollout` _string_ | Rollout specifies the template used for rollout commits. |
| `delete` _string_ | Delete specifies the template used for delete commits. |


### HTTPSCredential



HTTPSCredential specifies basic auth credentials

_Appears in:_
- [GitCredentials](#gitcredentials)

| Field | Description |
| --- | --- |
| `username` _string_ | Username is the basic auth user name |
| `password` _string_ | Password is the basic auth password |


### Logging



Logging specifies logging configuration.

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `devMode` _boolean_ | DevModeEnabled enables verbose logs and changes the logging format to be<br />more human readable. |
| `level` _[Level](#level)_ | Level sets the granularity of logging. |


### OIDCProvider



OIDCProvider specifies an OIDC provider.

_Appears in:_
- [SSO](#sso)

| Field | Description |
| --- | --- |
| `name` _string_ | Name is a human-readable name of the provider. If set this will be used<br />instead of the provider id (the key in<br />`PlatformConfig.Auth.SSO.OIDCProviders`) |
| `issuerURL` _string_ | IssuerURL is the URL for the OIDC issuer endpoint. |
| `clientID` _string_ | ClientID is the OAuth client ID. |
| `clientSecret` _string_ | ClientSecret is the OAuth client secret. |
| `allowedDomains` _string array_ | AllowedDomains is a list of email domains to allow. If left empty any<br />successful authentication on the provider is allowed. |
| `scopes` _string array_ | Scopes is a list of additional scopes other than `openid`, `email` and<br />`profile`. |
| `groupsClaim` _string_ | GroupsClaim is the path to a claim in the JWT containing a string or<br />list of strings of group names. |
| `disableJITGroups` _boolean_ | DisableJITGroups disables creation of groups found through OIDC in rig. |
| `groupMapping` _object (keys:string, values:string)_ | GroupMapping is a mapping from OIDC provided group names to group names<br />used in rig. If an OIDC provided group name is not provided in this<br />mapping we will use the OIDC provided groupname in rig. |
| `icon` _[OIDCProviderIcon](#oidcprovidericon)_ | Icon is what icon to show for this provider. |
| `disableUserMerging` _boolean_ | DisableUserMerging disallows merging their OIDC account with an existing user in rig.<br />This effectively means, that if a user is created using OIDC, then it can only login<br />using that OIDC provider. |


### OIDCProviderIcon

_Underlying type:_ _string_

OIDCProviderIcon is a string representing what provider icon should be shown
on the login page. Valid options: "google", "azure", "aws", "facebook",
"keycloak".

_Appears in:_
- [OIDCProvider](#oidcprovider)



### OperatorConfig



OperatorConfig is the Schema for the operator config API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `config.rig.dev/v1alpha1`
| `kind` _string_ | `OperatorConfig`
| `webhooksEnabled` _boolean_ | WebhooksEnabled sets wether or not webhooks should be enabled. When<br />enabled a certificate should be mounted at the webhook server<br />certificate path. Defaults to true if omitted. |
| `devModeEnabled` _boolean_ | DevModeEnabled enables verbose logs and changes the logging format to be<br />more human readable. |
| `leaderElectionEnabled` _boolean_ | LeaderElectionEnabled enables leader election when running multiple<br />instances of the operator. |
| `pipeline` _[Pipeline](#pipeline)_ | Pipeline defines the capsule controller pipeline |


### PathPrefixes



PathPrefixes is the (possibly templated) path prefix to commit to in git repository
depending on which resource is being written.

_Appears in:_
- [ClusterGit](#clustergit)

| Field | Description |
| --- | --- |
| `capsule` _string_ |  |
| `project` _string_ |  |


### Pipeline





_Appears in:_
- [OperatorConfig](#operatorconfig)

| Field | Description |
| --- | --- |
| `serviceAccountStep` _[CapsuleStep](#capsulestep)_ | How to handle the service account step of capsules in the cluster.<br />Defaults to rigdev.service_account. |
| `deploymentStep` _[CapsuleStep](#capsulestep)_ | How to handle the deployment step of capsules in the cluster.<br />Defaults to rigdev.deployment. |
| `routesStep` _[CapsuleStep](#capsulestep)_ | How to handle the routes for capsules in the cluster.<br />If left empty, routes will not be handled. |
| `cronJobsStep` _[CapsuleStep](#capsulestep)_ | How to handle the cronjob step of capsules in the cluster.<br />Defaults to rigdev.cron_jobs |
| `vpaStep` _[CapsuleStep](#capsulestep)_ | How to handle the VPA step of capsules in the cluster.<br />If left empty, no VPAs will be created. |
| `serviceMonitorStep` _[CapsuleStep](#capsulestep)_ | How to handle the service monitor step of capsules in the cluster.<br />If left empty, no service monitors will be created.<br />rigdev.service_monitor plugin spawns a Prometheus ServiceMonitor per capsule<br />for use with a Prometheus Operator stack. |
| `steps` _[Step](#step) array_ | Steps to perform as part of running the operator. |
| `customPlugins` _[CustomPlugin](#customplugin) array_ | CustomPlugins enables custom plugins to be injected into the<br />operator. The plugins injected here can then be referenced in 'steps' |
| `capsuleExtensions` _object (keys:string, values:[CapsuleStep](#capsulestep))_ | CapsuleExtensions supported by the Operator. Each extension supported<br />should be configured in the map, with an additional plugin name. |


### PlatformConfig



PlatformConfig is the Schema for the platform config API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `config.rig.dev/v1alpha1`
| `kind` _string_ | `PlatformConfig`
| `port` _integer_ | Port sets the port the platform should listen on |
| `publicURL` _string_ | PublicUrl sets the public url for the platform. This is used for<br />generating urls for the platform when using oauth2. |
| `telemetryEnabled` _boolean_ | TelemetryEnabled specifies wether or not we are allowed to collect usage<br />data. Defaults to true. |
| `auth` _[Auth](#auth)_ | Auth holds authentication configuration. |
| `client` _[Client](#client)_ | Client holds configuration for clients used in the platform. |
| `repository` _[Repository](#repository)_ | Repository specifies the type of db to use along with secret key |
| `cluster` _[Cluster](#cluster)_ | Cluster holds cluster specific configuration<br />Deprecated: Use `clusters` instead. |
| `email` _[Email](#email)_ | Email holds the default configuration for sending emails. Either using mailjet or using SMTP. |
| `logging` _[Logging](#logging)_ | Logging holds information about the granularity of logging |
| `clusters` _object (keys:string, values:[Cluster](#cluster))_ | Clusters the platform has access to. |
| `dockerRegistries` _object (keys:string, values:[DockerRegistryCredentials](#dockerregistrycredentials))_ | DockerRegistries holds configuration for multiple docker registries. The key is the host-prefix of the registry |
| `capsuleExtensions` _object (keys:string, values:[Extension](#extension))_ | CapsuleExtensions contains typed extensions to the Capsule spec. |


### Plugin





_Appears in:_
- [Step](#step)

| Field | Description |
| --- | --- |
| `tag` _string_ | Optional tag which is readable by plugin when executed |
| `name` _string_ | Name of the plugin to run.<br />Deprecated, use Plugin. |
| `plugin` _string_ | Name of the plugin to run. |
| `config` _string_ | Config is a string defining the plugin-specific configuration of the plugin. |




### Repository



Repository specifies repository configuration

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `store` _string_ | Store is what database will be used, can only be postgres. |
| `secret` _string_ | Secret is a secret key used for encrypting sensitive data before saving<br />it in the database. |


### SSHCredential



SSHCredential specifies SSH credentials

_Appears in:_
- [GitCredentials](#gitcredentials)

| Field | Description |
| --- | --- |
| `privateKey` _string_ | PrivateKey is a PEM encoded SSH private key. |
| `password` _string_ | PrivateKeyPassword is an optional password for the SSH private key. |


### SSO



SSO specifies single sign on configuration.

_Appears in:_
- [Auth](#auth)

| Field | Description |
| --- | --- |
| `oidcProviders` _object (keys:string, values:[OIDCProvider](#oidcprovider))_ | OIDCProviders specifies enabled OIDCProviders which can be used for<br />login. |


### Step





_Appears in:_
- [Pipeline](#pipeline)

| Field | Description |
| --- | --- |
| `tag` _string_ | Optional tag which is readable by plugins when executed |
| `match` _[CapsuleMatch](#capsulematch)_ | Match requirements for running the Step on a given Capsule. |
| `plugins` _[Plugin](#plugin) array_ | Plugins to run as part of this step. |
| `namespaces` _string array_ | If set, only capsules in one of the namespaces given will have this step run.<br />Deprecated, use Match.Namespaces. |
| `capsules` _string array_ | If set, only execute the plugin on the capsules specified.<br />Deprecated, use Match.Names. |
| `enableForPlatform` _boolean_ | If set, will enable the step for the Rig platform which is a Capsule as well<br />Deprecated, use Match.EnableForPlatform. |


### URLMatch

_Underlying type:_ _string_



_Appears in:_
- [GitAuth](#gitauth)







<hr class="solid" />


:::info generated from source code
This page is generated based on go source code. If you have suggestions for
improvements for this page, please open an issue at
[github.com/rigdev/rig](https://github.com/rigdev/rig/issues/new), or a pull
request with changes to [the go source
files](https://github.com/rigdev/rig/tree/main/pkg/api).
:::