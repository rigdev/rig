# API Reference

## Packages
- [config.rig.dev/v1alpha1](#configrigdevv1alpha1)


## config.rig.dev/v1alpha1

Package v1alpha1 contains API Schema definitions for the config v1alpha1 API group

### Resource Types
- [OperatorConfig](#operatorconfig)
- [PlatformConfig](#platformconfig)



#### Auth





_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `secret` _string_ |  |
| `certificateFile` _string_ |  |
| `certificateKeyFile` _string_ |  |


#### CertManagerConfig





_Appears in:_
- [OperatorConfig](#operatorconfig)

| Field | Description |
| --- | --- |
| `clusterIssuer` _string_ | ClusterIssuer to use for issueing ingress certificates |
| `createCertificateResources` _boolean_ | CreateCertificateResources specifies wether to create Certificate resources. If this is not enabled we will use ingress annotations. This is handy in environments where the ingress-shim isen't enabled. |


#### Client





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


#### ClientDocker





_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `host` _string_ |  |


#### ClientMailjet





_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `apiKey` _string_ |  |
| `secretKey` _string_ |  |


#### ClientMongo





_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `user` _string_ |  |
| `password` _string_ |  |
| `host` _string_ | Host of the mongo server. This is both the host and port. |


#### ClientOperator





_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `baseUrl` _string_ |  |


#### ClientPostgres





_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `user` _string_ |  |
| `password` _string_ |  |
| `host` _string_ |  |
| `port` _integer_ |  |
| `database` _string_ | Database in the postgres server to use |
| `insecure` _boolean_ | Use SSL when connecting to the postgres server |


#### ClientSMTP





_Appears in:_
- [Client](#client)

| Field | Description |
| --- | --- |
| `host` _string_ |  |
| `port` _integer_ |  |
| `username` _string_ |  |
| `password` _string_ |  |


#### Cluster





_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `type` _[ClusterType](#clustertype)_ | Type of the cluster - either docker or k8s |
| `devRegistry` _[DevRegistry](#devregistry)_ |  |
| `git` _[ClusterGit](#clustergit)_ |  |


#### ClusterGit





_Appears in:_
- [Cluster](#cluster)

| Field | Description |
| --- | --- |
| `url` _string_ |  |
| `branch` _string_ |  |
| `pathPrefix` _string_ |  |
| `credentials` _[GitCredentials](#gitcredentials)_ |  |
| `author` _[GitAuthor](#gitauthor)_ |  |
| `templates` _[GitTemplates](#gittemplates)_ |  |


#### ClusterType

_Underlying type:_ _string_



_Appears in:_
- [Cluster](#cluster)



#### DevRegistry





_Appears in:_
- [Cluster](#cluster)

| Field | Description |
| --- | --- |
| `host` _string_ |  |
| `clusterHost` _string_ |  |


#### Email





_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `from` _string_ |  |
| `type` _string_ |  |




#### GitAuthor





_Appears in:_
- [ClusterGit](#clustergit)

| Field | Description |
| --- | --- |
| `name` _string_ |  |
| `email` _string_ |  |


#### GitCredentials





_Appears in:_
- [ClusterGit](#clustergit)

| Field | Description |
| --- | --- |
| `https` _[HTTPSCredential](#httpscredential)_ |  |
| `ssh` _[SSHCredential](#sshcredential)_ |  |


#### GitTemplates





_Appears in:_
- [ClusterGit](#clustergit)

| Field | Description |
| --- | --- |
| `rollout` _string_ |  |
| `delete` _string_ |  |


#### HTTPSCredential





_Appears in:_
- [GitCredentials](#gitcredentials)

| Field | Description |
| --- | --- |
| `username` _string_ |  |
| `password` _string_ |  |


#### IngressConfig





_Appears in:_
- [OperatorConfig](#operatorconfig)

| Field | Description |
| --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations for all ingress resources created. |
| `className` _string_ | ClassName specifies the default ingress class to use for all ingress resources created. |


#### Logging





_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `devMode` _boolean_ | DevModeEnabled enables verbose logs and changes the logging format to be more human readable. |
| `level` _[Level](#level)_ | Level sets the granularity of logging |


#### OAuth





_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `google` _[OAuthClientCredentials](#oauthclientcredentials)_ |  |
| `github` _[OAuthClientCredentials](#oauthclientcredentials)_ |  |
| `facebook` _[OAuthClientCredentials](#oauthclientcredentials)_ |  |


#### OAuthClientCredentials





_Appears in:_
- [OAuth](#oauth)

| Field | Description |
| --- | --- |
| `clientId` _string_ |  |
| `clientSecret` _string_ |  |


#### OperatorConfig



OperatorConfig is the Schema for the configs API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `config.rig.dev/v1alpha1`
| `kind` _string_ | `OperatorConfig`
| `webhooksEnabled` _boolean_ | WebhooksEnabled set wether or not webhooks should be enabled. When enabled a certificate should be mounted at the webhook server certificate path. Defaults to true if omitted. |
| `devModeEnabled` _boolean_ | DevModeEnabled enables verbose logs and changes the logging format to be more human readable. |
| `leaderElectionEnabled` _boolean_ | LeaderElectionEnabled enables leader election when running multiple instances of the operator. |
| `certManager` _[CertManagerConfig](#certmanagerconfig)_ | Certmanager holds configuration for how the operator should create certificates for ingress resources. |
| `ingress` _[IngressConfig](#ingressconfig)_ | Ingress holds the configuration for ingress resources created by the operator. |


#### PlatformConfig



OperatorConfig is the Schema for the configs API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `config.rig.dev/v1alpha1`
| `kind` _string_ | `PlatformConfig`
| `port` _integer_ | Port sets the port the platform should listen on |
| `publicUrl` _string_ | PublicUrl sets the public url for the platform. This is used for generating urls for the platform when using oauth2. |
| `telemetryEnabled` _boolean_ |  |
| `auth` _[Auth](#auth)_ |  |
| `client` _[Client](#client)_ | Client holds configuration for clients used in the platform. |
| `repository` _[Repository](#repository)_ | Repository specifies the type of db to use along with secret key |
| `oauth` _[OAuth](#oauth)_ | OAuth holds configuration for oauth2 clients, namely google, github and facebook. |
| `cluster` _[Cluster](#cluster)_ |  |
| `email` _[Email](#email)_ | Email holds configuration for sending emails. Either using mailjet or using SMTP |
| `logging` _[Logging](#logging)_ | Loggin holds information about the granularity of logging |


#### Repository





_Appears in:_
- [PlatformConfig](#platformconfig)

| Field | Description |
| --- | --- |
| `store` _string_ | Type of db to use |
| `secret` _string_ |  |


#### SSHCredential





_Appears in:_
- [GitCredentials](#gitcredentials)

| Field | Description |
| --- | --- |
| `privateKey` _string_ |  |
| `password` _string_ |  |


