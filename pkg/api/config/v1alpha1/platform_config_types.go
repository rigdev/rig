package v1alpha1

import (
	"fmt"

	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func init() {
	SchemeBuilder.Register(&PlatformConfig{})
}

// PlatformConfig is the Schema for the platform config API
// +kubebuilder:object:root=true
type PlatformConfig struct {
	metav1.TypeMeta `json:",inline"`

	// Port sets the port the platform should listen on
	Port int `json:"port,omitempty"`

	// PublicUrl sets the public url for the platform. This is used for
	// generating urls for the platform when using oauth2.
	PublicURL string `json:"publicURL,omitempty"`

	// TelemetryEnabled specifies wether or not we are allowed to collect usage
	// data. Defaults to true.
	TelemetryEnabled bool `json:"telemetryEnabled,omitempty"`

	// Auth holds authentication configuration.
	Auth Auth `json:"auth,omitempty"`

	// Client holds configuration for clients used in the platform.
	Client Client `json:"client,omitempty"`

	// Repository specifies the type of db to use along with secret key
	Repository Repository `json:"repository,omitempty"`

	// Cluster holds cluster specific configuration
	// Deprecated: Use `clusters` instead.
	Cluster Cluster `json:"cluster,omitempty"`

	// Email holds the default configuration for sending emails. Either using mailjet or using SMTP.
	Email Email `json:"email,omitempty"`

	// Logging holds information about the granularity of logging
	Logging Logging `json:"logging,omitempty"`

	// Clusters the platform has access to.
	Clusters map[string]Cluster `json:"clusters,omitempty"`

	// DockerRegistries holds configuration for multiple docker registries. The key is the host-prefix of the registry
	DockerRegistries map[string]DockerRegistryCredentials `json:"dockerRegistries,omitempty"`
}

// Auth specifies authentication configuration.
type Auth struct {
	// Secret specifies a secret which will be used for jwt signatures.
	Secret string `json:"secret,omitempty"`

	// CertificateFile specifies a path to a PEM encoded certificate file which
	// will be used for validating jwt signatures.
	CertificateFile string `json:"certificateFile,omitempty"`

	// CertificateKeyFile specifies a path to a PEM encoded certificate key
	// which will be used for jwt signatures.
	CertificateKeyFile string `json:"certificateKeyFile,omitempty"`

	// DisablePasswords disables password authentication. This makes sense if
	// you want to require SSO, as login method.
	DisablePasswords bool `json:"disablePasswords,omitempty"`

	// SSO specifies single sign on configuration.
	SSO SSO `json:"sso,omitempty"`

	// AllowRegister specifies if users are allowed to register new accounts.
	AllowRegister bool `json:"allowRegister,omitempty"`

	// IsVerified specifies if users are required to verify their email address.
	RequireVerification bool `json:"requireVerification,omitempty"`

	// SendWelcomeEmail specifies if a welcome email should be sent to new users.
	// This will use the default email config
	SendWelcomeEmail bool `json:"sendWelcomeEmail,omitempty"`
}

// SSO specifies single sign on configuration.
type SSO struct {
	// OIDCProviders specifies enabled OIDCProviders which can be used for
	// login.
	OIDCProviders map[string]OIDCProvider `json:"oidcProviders,omitempty"`
}

// OIDCProvider specifies an OIDC provider.
type OIDCProvider struct {
	// Name is a human-readable name of the provider. If set this will be used
	// instead of the provider id (the key in
	// `PlatformConfig.Auth.SSO.OIDCProviders`)
	Name string `json:"name,omitempty"`

	// IssuerURL is the URL for the OIDC issuer endpoint.
	IssuerURL string `json:"issuerURL,omitempty"`

	// ClientID is the OAuth client ID.
	ClientID string `json:"clientID,omitempty"`

	// ClientSecret is the OAuth client secret.
	ClientSecret string `json:"clientSecret,omitempty"`

	// AllowedDomains is a list of email domains to allow. If left empty any
	// successful authentication on the provider is allowed.
	AllowedDomains []string `json:"allowedDomains,omitempty"`

	// Scopes is a list of additional scopes other than `openid`, `email` and
	// `profile`.
	Scopes []string `json:"scopes"`

	// GroupsClaim is the path to a claim in the JWT containing a string or
	// list of strings of group names.
	GroupsClaim string `json:"groupsClaim,omitempty"`

	// DisableJITGroups disables creation of groups found through OIDC in rig.
	DisableJITGroups *bool `json:"disableJITGroups,omitempty"`

	// GroupMapping is a mapping from OIDC provided group names to group names
	// used in rig. If an OIDC provided group name is not provided in this
	// mapping we will use the OIDC provided groupname in rig.
	GroupMapping map[string]string `json:"groupMapping,omitempty"`

	// Icon is what icon to show for this provider.
	Icon OIDCProviderIcon `json:"icon,omitempty"`

	// DisableUserMerging disallows merging their OIDC account with an existing user in rig.
	// This effectively means, that if a user is created using OIDC, then it can only login
	// using that OIDC provider.
	DisableUserMerging *bool `json:"disableUserMerging,omitempty"`
}

// OIDCProviderIcon is a string representing what provider icon should be shown
// on the login page. Valid options: "google", "azure", "aws", "facebook",
// "keycloak".
type OIDCProviderIcon string

const (
	OIDCProviderIconGoogle   OIDCProviderIcon = "google"
	OIDCProviderIconAzure    OIDCProviderIcon = "azure"
	OIDCProviderIconAWS      OIDCProviderIcon = "aws"
	OIDCProviderIconFacebook OIDCProviderIcon = "facebook"
	OIDCProviderIconKeycloak OIDCProviderIcon = "keycloak"
)

// Client holds various client configuration
type Client struct {
	// Postgres holds configuration for the postgres client.
	Postgres ClientPostgres `json:"postgres,omitempty"`

	// Docker sets the host for the Docker client.
	Docker ClientDocker `json:"docker,omitempty"`

	// Deprecated: use 'client.mailjets' instead.
	// Mailjet sets the API key and secret for the Mailjet client.
	Mailjet ClientMailjet `json:"mailjet,omitempty"`

	// Mailjets holds configuration for multiple mailjet clients.
	// The key is the id of the client, which should be unique across Mailjet and SMTP clients.
	Mailjets map[string]ClientMailjet `json:"mailjets,omitempty"`

	// Deprecated: use 'client.smtps' instead.
	// SMTP sets the host, port, username and password for the SMTP client.
	SMTP ClientSMTP `json:"smtp,omitempty"`

	// SMTPs holds configuration for muliple SMTP clients.
	// The key is the id of the client, which should be unique across Mailjet and SMTP clients.
	SMTPs map[string]ClientSMTP `json:"smtps,omitempty"`

	// Operator sets the base url for the Operator client.
	Operator ClientOperator `json:"operator,omitempty"`

	// Slack holds configuration for sending slack messages. The key is the id of the client.
	// For example the workspace in which the app is installed
	Slack map[string]ClientSlack `json:"slack,omitempty"`

	// Git client configuration for communicating with multiple repositories.
	Git ClientGit `json:"git,omitempty"`
}

// Logging specifies logging configuration.
type Logging struct {
	// DevModeEnabled enables verbose logs and changes the logging format to be
	// more human readable.
	DevMode bool `json:"devMode,omitempty"`

	// Level sets the granularity of logging.
	Level zapcore.Level `json:"level,omitempty"`
}

// ClientPostgres specifies the configuration for the postgres client.
type ClientPostgres struct {
	// User is the database user used when connecting to the postgres database.
	User string `json:"user,omitempty"`

	// Password is the password used when connecting to the postgres database.
	Password string `json:"password,omitempty"`

	// Host is the host where the postgres database can be reached.
	Host string `json:"host,omitempty"`

	// Port is the port of the postgres database server.
	Port int `json:"port,omitempty"`

	// Database in the postgres server to use
	Database string `json:"database,omitempty"`

	// Insecure is wether to use SSL when connecting to the postgres server
	Insecure bool `json:"insecure,omitempty"`
}

// ClientDocker specifies the configuration for the docker client.
type ClientDocker struct {
	// Host where the docker daemon can be reached.
	Host string `json:"host,omitempty"`
}

type ClientSlack struct {
	// Slack authentication token.
	Token string `json:"token,omitempty"`
}

// ClientMailjet specifes the configuration for the mailjet client.
type ClientMailjet struct {
	// APIKey is the mailjet API key
	APIKey string `json:"apiKey,omitempty"`

	// SecretKey is the mailjet secret key
	SecretKey string `json:"secretKey,omitempty"`
}

// ClientSMTP specifies the configuration for the SMTP client.
type ClientSMTP struct {
	// Host is the SMTP server host.
	Host string `json:"host,omitempty"`

	// Port is the SMTP server port to use.
	Port int `json:"port,omitempty"`

	// Username used when connecting to the SMTP server.
	Username string `json:"username,omitempty"`

	// Password used when connecting to the SMTP server.
	Password string `json:"password,omitempty"`
}

// ClientOperator specifies the configuration for the operator client.
type ClientOperator struct {
	// BaseURL is the URL used to connect to the operator API
	BaseURL string `json:"baseUrl,omitempty"`
}

// Repository specifies repository configuration
type Repository struct {
	// Store is what database will be used, can only be postgres.
	Store string `json:"store,omitempty"`

	// Secret is a secret key used for encrypting sensitive data before saving
	// it in the database.
	Secret string `json:"secret,omitempty"`
}

// Cluster specifies cluster configuration
type Cluster struct {
	// URL to communicate to the cluster. If set, a Token and CertificateAuthority should
	// be provided as well.
	// If not set, the cluster is interpreted to be the cluster in which the platform runs.
	URL string `json:"url,omitempty"`

	// Token for communicating with the cluster. Available through a service-account's secret.
	Token string `json:"token,omitempty"`

	// Certificate authority for communicating with the cluster. Available through a service-account's secret.
	CertificateAuthority string `json:"certificateAuthority,omitempty"`

	// Type of the cluster - either `docker` or `k8s`.
	Type ClusterType `json:"type,omitempty"`

	// DevRegistry configuration
	DevRegistry DevRegistry `json:"devRegistry,omitempty"`

	// Git sets up gitops write back for this cluster.
	Git ClusterGit `json:"git,omitempty"`

	// If set, secrets will be created if needed, for pulling images.
	CreatePullSecrets *bool `json:"createPullSecrets,omitempty"`
}

// ClusterGit specifies configuration for git integration. This can be used to
// tie rig into a gitops setup.
type ClusterGit struct {
	// URL is the git repository URL.
	URL string `json:"url,omitempty"`

	// Branch to commit changes to.
	Branch string `json:"branch,omitempty"`

	// PathPrefix path to commit to in git repository.
	// Deprecated: Use `pathPrefixes` instead.
	PathPrefix string `json:"pathPrefix,omitempty"`

	// PathPrefixes path to commit to in git repository
	PathPrefixes PathPrefixes `json:"pathPrefixes,omitempty"`

	// Templates used for commit messages.
	Templates GitTemplates `json:"templates,omitempty"`

	// Credentials to use when connecting to git.
	// Deprecated: Use `client.git.auths` instead.
	Credentials GitCredentials `json:"credentials,omitempty"`

	// Author used when creating commits.
	// Deprecated: Use `client.git.author` instead.
	Author GitAuthor `json:"author,omitempty"`
}

// PathPrefixes is the (possibly templated) path prefix to commit to in git repository
// depending on which resource is being written.
type PathPrefixes struct {
	Capsule string `json:"capsule,omitempty"`
	Project string `json:"project,omitempty"`
}

// ClientGit contains configuration for git integrations.
// A given git repository can have authentication from either Auths or GitHubAuths with preference
// for GitHubAuths if there is a match.
type ClientGit struct {
	// Auths the git client can behave as.
	Auths []GitAuth `json:"auths,omitempty"`

	// GitHubAuths is authentication information for GitHub repositories
	GitHubAuths []GitHub `json:"gitHubAuths,omitempty"`

	GiLabAuths []GitLab `json:"gitLabAuths,omitempty"`

	// Author used when creating commits.
	Author GitAuthor `json:"author,omitempty"`
}

type GitAuth struct {
	// URL is a exact match for the repo-url this auth can be used for.
	URL string `json:"url,omitempty"`

	// URLPrefix is a prefix-match for the repo urls this auth can be used for.
	URLPrefix string `json:"urlPrefix,omitempty"`

	// Credentials to use when connecting to git.
	Credentials GitCredentials `json:"credentials,omitempty"`

	// If no web hook is confugured, pull the git repository at the set interval instead
	// to fetch changes. Defaults to 3 mins if no value.
	PullingIntervalSeconds int `json:"pullingIntervalSeconds"`
}

// GitHub contains configuration specifically for GitHub repositories.
// To enable pull requests on a GitHub repository, you must add GitHub authentication
// using appID, installationID and privateKey for a GitHub app with read/write access to
// pull requests.
// To have normal read/write access to a repository, you can forego GitHub app authentication
// if there is a GitAuth section with credentials for the given repository instead.
// If you have GitHub app authentication for a GitHub app with read/write access to the repository,
// you don't need a matching GitAuth section.
type GitHub struct {
	// Organization is the GitHub organization to match.
	Organization string `json:"organization"`

	// Repository matches the GitHub repository. If empty, matches all.
	Repository string `json:"repository,omitempty"`

	// Auth contains GitHub specific authentication configuration.
	Auth GitHubAuth `json:"auth,omitempty"`

	// Polling contains GitHub specific configuration.
	Polling GitHubPolling `json:"polling,omitempty"`
}

// GitHubAuth contains authentication information specifically for a GitHub repository.
// Authentication is done using GitHub apps. See https://docs.rig.dev/operator-manual/gitops#github-authentication
// for a guide on how to set it up.
type GitHubAuth struct {
	// AppID is the app ID of the GitHub app
	AppID int64 `json:"appID,omitempty"`

	// InstallationID is the installation ID of the GitHub app
	InstallationID int64 `json:"installationID,omitempty"`

	// PrivateKey is a PEM encoded SSH private key.
	PrivateKey string `json:"privateKey,omitempty"`

	// PrivateKeyPassword is an optional password for the SSH private key.
	PrivateKeyPassword string `json:"privateKeyPassword,omitempty"`
}

// GitHubPolling defines webhook/pulling configuration for a GitHub repository.
type GitHubPolling struct {
	// WebHookSecret is the secret used to validate incoming webhooks.
	WebhookSecret string `json:"webhookSecret,omitempty"`

	// If webHookSecret isn't set, pull the git repository at the set interval instead
	// to fetch changes. Defaults to 3 mins if no value.
	PullingIntervalSeconds int `json:"pullingIntervalSeconds,omitempty"`
}

// GitLab contains configuration specifically for GitLab repositories.
// To enable pull requests on a GitLab repository, you must add GitLab authentication
// using an access token.
// To have normal read/write access to a repository, you can forego GitLab access tokens
// if there is a GitAuth section with credentials for the given repository instead.
// If you have GitLab authentication for a repository, you don't need a matching GitAuth section.
type GitLab struct {
	// Groups is a sequence of GitLab groups.
	// The first is the main group and the rest a nesting of subgroups.
	// If Project is empty, the configuration will match any
	// GitLab repository whose (group, subgroups) sequence where 'groups' is a prefix.
	Groups []string `json:"groups,omitempty"`

	// Project is the GitLab project of the repository. Can be empty for matching all project names.
	Project string `json:"project,omitempty"`

	// Auth contains GitLab specific authentication configuration.
	Auth GitLabAuth `json:"auth,omitempty"`

	// Polling contains GitLab specific configuration.
	Polling GitLabPolling `json:"polling,omitempty"`
}

// GitLabAuth contains authentication information specifically for a GitLab repository.
// Authentication is done using an access token. See https://docs.rig.dev/operator-manual/gitops#gitlab-authentication
// for a guide on how to set it up.
type GitLabAuth struct {
	// AccessToken is an accessToken which is used to authenticate against the GitLab repository.
	Accesstoken string `json:"accessToken,omitempty"`
}

// GitLabPolling defines webhook/pulling configuration for a GitLab repository.
type GitLabPolling struct {
	// WebHookSecret is the secret used to validate incoming webhooks.
	WebhookSecret string `json:"webhookSecret,omitempty"`

	// If webHookSecret isn't set, pull the git repository at the set interval instead
	// to fetch changes. Defaults to 3 mins if no value.
	PullingIntervalSeconds int `json:"pullingIntervalSeconds,omitempty"`
}

// GitCredentials specifies how to authenticate against git.
type GitCredentials struct {
	// HTTPS specifies basic auth credentials.
	HTTPS HTTPSCredential `json:"https,omitempty"`

	// SSH specifies SSH credentials.
	SSH SSHCredential `json:"ssh,omitempty"`
}

// HTTPSCredential specifies basic auth credentials
type HTTPSCredential struct {
	// Username is the basic auth user name
	Username string `json:"username,omitempty"`

	// Password is the basic auth password
	Password string `json:"password,omitempty"`
}

// SSHCredential specifies SSH credentials
type SSHCredential struct {
	// PrivateKey is a PEM encoded SSH private key.
	PrivateKey string `json:"privateKey,omitempty"`

	// PrivateKeyPassword is an optional password for the SSH private key.
	PrivateKeyPassword string `json:"password,omitempty"`
}

// GitAuthor specifies a git commit author
type GitAuthor struct {
	// Name of author
	Name string `json:"name,omitempty"`

	// Email of author
	Email string `json:"email,omitempty"`
}

// GitTemplates specifies the templates used for creating commits.
type GitTemplates struct {
	// Rollout specifies the template used for rollout commits.
	Rollout string `json:"rollout,omitempty"`

	// Delete specifies the template used for delete commits.
	Delete string `json:"delete,omitempty"`
}

// DevRegistry specifies configuration for the dev registry support.
type DevRegistry struct {
	// Host is the host used in image names when pushing to the registry from
	// outside of the cluster.
	Host string `json:"host,omitempty"`

	// ClusterHost is the host where the registry can be reached from within
	// the cluster. Any image which is named after `Host` will be rename to use
	// `ClusterHost` instead. This ensures that the image can be pulled from
	// within the cluster.
	ClusterHost string `json:"clusterHost,omitempty"`
}

type DockerRegistryCredentials struct {
	// Username for the docker registry.
	Username string `json:"username,omitempty"`
	// Password for the docker registry.
	Password string `json:"password,omitempty"`
	// Script (shell) to execute that should echo the credentials.
	// The output is expected to be a single line (with new-line termination) of format `<username>:<password>`.
	Script string `json:"script,omitempty"`
	// Expire is the maximum duration a credential will be cached for, before it's recycled.
	// If a cached credential is rejected before this time, it may be renewed before this duration is expired.
	// Default is `12h`.
	Expire *metav1.Duration `json:"expire,omitempty"`
}

// ClusterType is a cluster type.
type ClusterType string

const (
	// ClusterTypeDocker is the docker cluster type.
	ClusterTypeDocker ClusterType = "docker"
	// ClusterTypeKubernetes is the kubernetes cluster type.
	ClusterTypeKubernetes ClusterType = "k8s"
)

// Email holds configuration for sending emails. Either using mailjet or using SMTP
type Email struct {
	// ID is the specified id an email configuration.
	ID string `json:"id,omitempty"`

	// From is who is set as the sender of rig emails.
	From string `json:"from,omitempty"`

	// Deprecated: ID for an email configuration is used instead.
	Type EmailType `json:"type,omitempty"`
}

// EmailType represents a type of mailing provider
type EmailType string

const (
	// EmailTypeNoEmail disables mail sending.
	EmailTypeNoEmail = ""
	// EmailTypeMailjet uses the mailjet API for sending emails.
	EmailTypeMailjet = "mailjet"
	// EmailTypeSMTP uses regular SMTP for sending emails.
	EmailTypeSMTP = "smtp"
)

func NewDefaultPlatform() *PlatformConfig {
	cfg := &PlatformConfig{
		Port:      4747,
		PublicURL: "",
		Logging: Logging{
			DevMode: false,
			Level:   zapcore.InfoLevel,
		},
		TelemetryEnabled: true,
		Auth: Auth{
			Secret:             "",
			CertificateFile:    "",
			CertificateKeyFile: "",
		},
		Client: Client{
			Postgres: ClientPostgres{
				User:     "",
				Password: "",
				Host:     "",
				Port:     5432,
				Database: "rig",
				Insecure: false,
			},
			Docker: ClientDocker{
				Host: "",
			},
			Mailjets: map[string]ClientMailjet{},
			SMTPs:    map[string]ClientSMTP{},
			Operator: ClientOperator{
				BaseURL: "rig-operator:9000",
			},
		},
		Repository: Repository{
			Store:  "postgres",
			Secret: "",
		},
		Email: Email{
			ID:   "",
			From: "",
		},
		DockerRegistries: map[string]DockerRegistryCredentials{},
	}

	cfg.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "config.rig.dev",
		Version: "v1alpha1",
		Kind:    "PlatformConfig",
	})

	return cfg
}

func (cfg *PlatformConfig) Validate() error {
	if cfg.Cluster.Type != "" && len(cfg.Clusters) != 0 {
		return fmt.Errorf("only one of `cluster` and `clusters` must be set")
	}

	var errs field.ErrorList
	errs = append(errs, cfg.Cluster.validate(field.NewPath("clusters"))...)

	return errs.ToAggregate()
}

func (c Cluster) validate(path *field.Path) field.ErrorList {
	return c.Git.validate(path.Child("git"))
}

func (g ClusterGit) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if g.PathPrefix != "" && g.PathPrefixes != (PathPrefixes{}) {
		return append(errs, field.Invalid(path, g, "can't set both `pathPrefix` and `pathPrefixes`"))
	}

	return errs
}

func (cfg *PlatformConfig) Migrate() {
	for _, c := range cfg.Clusters {
		if c.Git.URL != "" && c.Git.Credentials != (GitCredentials{}) {
			cfg.Client.Git.Auths = append(cfg.Client.Git.Auths, GitAuth{
				URL:         c.Git.URL,
				Credentials: c.Git.Credentials,
			})
		}
		if cfg.Client.Git.Author.Name == "" {
			cfg.Client.Git.Author.Name = c.Git.Author.Name
			cfg.Client.Git.Author.Email = c.Git.Author.Email
		}
	}

	if cfg.Client.Mailjet != (ClientMailjet{}) {
		cfg.Client.Mailjets["mailjet"] = cfg.Client.Mailjet
	}

	if cfg.Client.SMTP != (ClientSMTP{}) {
		cfg.Client.SMTPs["smtp"] = cfg.Client.SMTP
	}

	if cfg.Email != (Email{}) && cfg.Email.Type != "" {
		switch cfg.Email.Type {
		case EmailTypeMailjet:
			if cfg.Email.ID == "" {
				cfg.Email.ID = "mailjet"
			}
		case EmailTypeSMTP:
			if cfg.Email.ID == "" {
				cfg.Email.ID = "smtp"
			}
		}
	}
}
