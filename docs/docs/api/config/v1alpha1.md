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
| `certificateFile` _string_ | CertificateFile specifies a path to a PEM encoded certificate file which will be used for validating jwt signatures. |
| `certificateKeyFile` _string_ | CertificateKeyFile specifies a path to a PEM encoded certificate key which will be used for jwt signatures. |
| `disablePasswords` _boolean_ | DisablePasswords disables password authentication. This makes sense if you want to require SSO, as login method. |
| `sso` _[SSO](#sso)_ | SSO specifies single sign on configuration. |


### CertManagerConfig





_Appears in:_
- [OperatorConfig](#operatorconfig)

| Field | Description |
| --- | --- |
| `clusterIssuer` _string_ | ClusterIssuer to use for issueing ingress certificates |
| `createCertificateResources` _boolean_ | CreateCertificateResources specifies wether to create Certificate resources. If this is not enabled we will use ingress annotations. This is handy in environments where the ingress-shim isn't enabled. |


### Client



Client holds various client configuration

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `postgres` _[ClientPostgres](#clientpostgres)_ | Postgres holds configuration for the postgres client. |
| `mongo` _[ClientMongo](#clientmongo)_ | Mongo holds configuration for the Mongo client. |
| `docker` _[ClientDocker](#clientdocker)_ | Docker sets the host for the Docker client. |
| `mailjet` _[ClientMailjet](#clientmailjet)_ | Mailjet sets the API key and secret for the Mailjet client. |
| `smtp` _[ClientSMTP](#clientsmtp)_ | SMTP sets the host, port, username and password for the SMTP client. |
| `operator` _[ClientOperator](#clientoperator)_ | Operator sets the base url for the Operator client. |


### ClientDocker



ClientDocker specifies the configuration for the docker client.

_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `host` _string_ | Host where the docker daemon can be reached. |


### ClientMailjet



ClientMailjet specifes the configuration for the mailjet client.

_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `apiKey` _string_ | APIKey is the mailjet API key |
| `secretKey` _string_ | SecretKey is the mailjet secret key |


### ClientMongo



ClientMongo specifies the configuration for the mongo client.

_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `user` _string_ | User is the database user used when connecting to the mongodb server. |
| `password` _string_ | Password is used when connecting to the mongodb server. |
| `host` _string_ | Host of the mongo server. This is both the host and port. |


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


### Cluster



Cluster specifies cluster configuration

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `url` _string_ | URL to communicate to the cluster. If set, a Token and CertificateAuthority should be provided as well. |
| `token` _string_ | Token for communicating with the cluster. Available through a service-account's secret. |
| `certificateAuthority` _string_ | Certificate authority for communicating with the cluster. Available through a service-account's secret. |
| `type` _[ClusterType](#clustertype)_ | Type of the cluster - either `docker` or `k8s`. |
| `devRegistry` _[DevRegistry](#devregistry)_ | DevRegistry configuration |
| `git` _[ClusterGit](#clustergit)_ | Git sets up gitops write back for this cluster. |


### ClusterGit



ClusterGit specifies configuration for git integration. This can be used to tie rig into a gitops setup.

_Appears in:_
- [Cluster](#cluster)

| Field | Description |
| --- | --- |
| `url` _string_ | URL is the git repository URL. |
| `branch` _string_ | Branch to commit changes to. |
| `pathPrefix` _string_ | PathPrefix path to commit to in git repository. |
| `credentials` _[GitCredentials](#gitcredentials)_ | Credentials to use when connecting to git. |
| `author` _[GitAuthor](#gitauthor)_ | Author used when creating commits. |
| `templates` _[GitTemplates](#gittemplates)_ | Templates used for commit messages. |


### ClusterType

_Underlying type:_ _string_

ClusterType is a cluster type.

_Appears in:_
- [Cluster](#cluster)



### DevRegistry



DevRegistry specifies configuration for the dev registry support.

_Appears in:_
- [Cluster](#cluster)

| Field | Description |
| --- | --- |
| `host` _string_ | Host is the host used in image names when pushing to the registry from outside of the cluster. |
| `clusterHost` _string_ | ClusterHost is the host where the registry can be reached from within the cluster. Any image which is named after `Host` will be rename to use `ClusterHost` instead. This ensures that the image can be pulled from within the cluster. |


### Email



Email holds configuration for sending emails. Either using mailjet or using SMTP

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `from` _string_ | From is who is set as the sender of rig emails. |
| `type` _[EmailType](#emailtype)_ | Type is what client rig should use to send emails. |


### EmailType

_Underlying type:_ _string_

EmailType represents a type of mailing provider

_Appears in:_
- [Email](#email)



### Environment



Environment configuration of a single environment.

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `cluster` _string_ | Cluster name the environment is hosted in. |
| `namespace_template` _string_ | NamespaceTemplate is used to generate the namespace name when configuring resources. Default is to set the namespace equal to the project name. |
| `default` _boolean_ | Default is true if this environment should be preferred for per-environment operations. |


### GitAuthor



GitAuthor specifies a git commit author

_Appears in:_
- [ClusterGit](#clustergit)

| Field | Description |
| --- | --- |
| `name` _string_ | Name of author |
| `email` _string_ | Email of author |


### GitCredentials



GitCredentials specifies how to authenticate against git.

_Appears in:_
- [ClusterGit](#clustergit)

| Field | Description |
| --- | --- |
| `https` _[HTTPSCredential](#httpscredential)_ | HTTPS specifies basic auth credentials. |
| `ssh` _[SSHCredential](#sshcredential)_ | SSH specifies SSH credentials. |


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


### IngressConfig





_Appears in:_
- [OperatorConfig](#operatorconfig)

| Field | Description |
| --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations for all ingress resources created. |
| `className` _string_ | ClassName specifies the default ingress class to use for all ingress resources created. |
| `pathType` _[PathType](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#pathtype-v1-networking)_ | PathType defines how ingress paths should be interpreted. Allowed values: Exact, Prefix, ImplementationSpecific |
| `disableTLS` _boolean_ | DisableTLS for ingress resources generated. This is useful if a 3rd-party component is handling the HTTPS TLS termination and certificates. |


### Logging



Logging specifies logging configuration.

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `devMode` _boolean_ | DevModeEnabled enables verbose logs and changes the logging format to be more human readable. |
| `level` _[Level](#level)_ | Level sets the granularity of logging. |


### OIDCProvider



OIDCProvider specifies an OIDC provider.

_Appears in:_
- [SSO](#sso)

| Field | Description |
| --- | --- |
| `name` _string_ | Name is a human-readable name of the provider. If set this will be used instead of the provider id (the key in `PlatformConfig.Auth.SSO.OIDCProviders`) |
| `issuerURL` _string_ | IssuerURL is the URL for the OIDC issuer endpoint. |
| `clientID` _string_ | ClientID is the OAuth client ID. |
| `clientSecret` _string_ | ClientSecret is the OAuth client secret. |
| `allowedDomains` _string array_ | AllowedDomains is a list of email domains to allow. If left empty any successful authentication on the provider is allowed. |
| `scopes` _string array_ | Scopes is a list of additional scopes other than `openid`, `email` and `profile`. |
| `groupsClaim` _string_ | GroupsClaim is the path to a claim in the JWT containing a string or list of strings of group names. |
| `disableJITGroups` _boolean_ | DisableJITGroups disables creation of groups found through OIDC in rig. |
| `groupMapping` _object (keys:string, values:string)_ | GroupMapping is a mapping from OIDC provided group names to group names used in rig. If an OIDC provided group name is not provided in this mapping we will use the OIDC provided groupname in rig. |
| `icon` _[OIDCProviderIcon](#oidcprovidericon)_ | Icon is what icon to show for this provider. |
| `disableUserMerging` _boolean_ | DisableUserMerging disallows merging their OIDC account with an existing user in rig. This effectively means, that if a user is created using OIDC, then it can only login using that OIDC provider. |


### OIDCProviderIcon

_Underlying type:_ _string_

OIDCProviderIcon is a string representing what provider icon should be shown on the login page. Valid options: "google", "azure", "aws", "facebook", "keycloak".

_Appears in:_
- [OIDCProvider](#oidcprovider)



### OperatorConfig



OperatorConfig is the Schema for the operator config API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `config.rig.dev/v1alpha1`
| `kind` _string_ | `OperatorConfig`
| `webhooksEnabled` _boolean_ | WebhooksEnabled sets wether or not webhooks should be enabled. When enabled a certificate should be mounted at the webhook server certificate path. Defaults to true if omitted. |
| `devModeEnabled` _boolean_ | DevModeEnabled enables verbose logs and changes the logging format to be more human readable. |
| `leaderElectionEnabled` _boolean_ | LeaderElectionEnabled enables leader election when running multiple instances of the operator. |
| `certManager` _[CertManagerConfig](#certmanagerconfig)_ | Certmanager holds configuration for how the operator should create certificates for ingress resources. |
| `service` _[ServiceConfig](#serviceconfig)_ | Service holds the configuration for service resources created by the operator. |
| `ingress` _[IngressConfig](#ingressconfig)_ | Ingress holds the configuration for ingress resources created by the operator. |
| `prometheusServiceMonitor` _[PrometheusServiceMonitor](#prometheusservicemonitor)_ | PrometheusServiceMonitor defines if Rig should spawn a Prometheus ServiceMonitor per capsule for use with a Prometheus Operator stack. |
| `verticalPodAutoscaler` _[VerticalPodAutoscaler](#verticalpodautoscaler)_ | VerticalPodAutoscaler holds the configuration for the VerticalPodAutoscaler resources potentially generated by the operator. |
| `steps` _[Step](#step) array_ | Steps to perform as part of running the operator. |


### PlatformConfig



PlatformConfig is the Schema for the platform config API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `config.rig.dev/v1alpha1`
| `kind` _string_ | `PlatformConfig`
| `port` _integer_ | Port sets the port the platform should listen on |
| `publicURL` _string_ | PublicUrl sets the public url for the platform. This is used for generating urls for the platform when using oauth2. |
| `telemetryEnabled` _boolean_ | TelemetryEnabled specifies wether or not we are allowed to collect usage data. Defaults to true. |
| `auth` _[Auth](#auth)_ | Auth holds authentication configuration. |
| `client` _[Client](#client)_ | Client holds configuration for clients used in the platform. |
| `repository` _[Repository](#repository)_ | Repository specifies the type of db to use along with secret key |
| `cluster` _[Cluster](#cluster)_ | Cluster holds cluster specific configuration Deprecated: Use `clusters` instead. |
| `email` _[Email](#email)_ | Email holds configuration for sending emails. Either using mailjet or using SMTP |
| `logging` _[Logging](#logging)_ | Logging holds information about the granularity of logging |
| `clusters` _object (keys:string, values:[Cluster](#cluster))_ | Clusters the platform has access to. |
| `environments` _object (keys:string, values:[Environment](#environment))_ | Environments of the platform. Each environment is backed by a cluster (allowing multi-tenant setups). |


### PrometheusServiceMonitor





_Appears in:_
- [OperatorConfig](#operatorconfig)

| Field | Description |
| --- | --- |
| `path` _string_ | Path is the path which Prometheus should query on ports. Defaults to /metrics if not set. |
| `portName` _string_ | PortName is the name of the port which Prometheus will query metrics on |


### Repository



Repository specifies repository configuration

_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `store` _string_ | Store is what database will be used can be either postgres or mongodb. |
| `secret` _string_ | Secret is a secret key used for encrypting sensitive data before saving it in the database. |


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
| `oidcProviders` _object (keys:string, values:[OIDCProvider](#oidcprovider))_ | OIDCProviders specifies enabled OIDCProviders which can be used for login. |


### ServiceConfig





_Appears in:_
- [OperatorConfig](#operatorconfig)

| Field | Description |
| --- | --- |
| `type` _[ServiceType](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#servicetype-v1-core)_ | Type of the service to generate. By default, services are of type ClusterIP. Valid values are ClusterIP, NodePort. |


### Step





_Appears in:_
- [OperatorConfig](#operatorconfig)

| Field | Description |
| --- | --- |
| `plugin` _string_ | Plugin to use in the current step. |
| `namespaces` _string array_ | If set, only capsules in one of the namespaces given will have this step run. |
| `capsules` _string array_ | If set, only execute the plugin on the capsules specified |
| `config` _string_ | Config is a string expected to be YAML defining the configuration for the plugin |


### VerticalPodAutoscaler





_Appears in:_
- [OperatorConfig](#operatorconfig)

| Field | Description |
| --- | --- |
| `enabled` _boolean_ | Enabled enables the creation of a VerticalPodAutoscaler per capsule |




<hr class="solid" />


:::info generated from source code
This page is generated based on go source code. If you have suggestions for
improvements for this page, please open an issue at
[github.com/rigdev/rig](https://github.com/rigdev/rig/issues/new), or a pull
request with changes to [the go source
files](https://github.com/rigdev/rig/tree/main/pkg/api).
:::