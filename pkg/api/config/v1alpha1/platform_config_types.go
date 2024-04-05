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

	// Email holds configuration for sending emails. Either using mailjet or using SMTP
	Email Email `json:"email,omitempty"`

	// Logging holds information about the granularity of logging
	Logging Logging `json:"logging,omitempty"`

	// Clusters the platform has access to.
	Clusters map[string]Cluster `json:"clusters,omitempty"`

	// Environments of the platform. Each environment is backed by a cluster (allowing multi-tenant setups).
	Environments map[string]Environment `json:"environments,omitempty"`
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

	// Mailjet sets the API key and secret for the Mailjet client.
	Mailjet ClientMailjet `json:"mailjet,omitempty"`

	// SMTP sets the host, port, username and password for the SMTP client.
	SMTP ClientSMTP `json:"smtp,omitempty"`

	// Operator sets the base url for the Operator client.
	Operator ClientOperator `json:"operator,omitempty"`
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
}

// Environment configuration of a single environment.
type Environment struct {
	// Cluster name the environment is hosted in.
	Cluster string `json:"cluster,omitempty"`

	// NamespaceTemplate is used to generate the namespace name when configuring resources.
	// Default is to set the namespace equal to the project name.
	NamespaceTemplate string `json:"namespace_template,omitempty"`

	// Default is true if this environment should be preferred for per-environment operations.
	Default bool `json:"default,omitempty"`
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

	// Credentials to use when connecting to git.
	Credentials GitCredentials `json:"credentials,omitempty"`

	// Author used when creating commits.
	Author GitAuthor `json:"author,omitempty"`

	// Templates used for commit messages.
	Templates GitTemplates `json:"templates,omitempty"`
}

// PathPrefixes is the (possibly templated) path prefix to commit to in git repository
// depending on which resource is being written.
type PathPrefixes struct {
	Capsule string `json:"capsule,omitempty"`
	Project string `json:"project,omitempty"`
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
	// From is who is set as the sender of rig emails.
	From string `json:"from,omitempty"`

	// Type is what client rig should use to send emails.
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
			Mailjet: ClientMailjet{
				APIKey:    "",
				SecretKey: "",
			},
			Operator: ClientOperator{
				BaseURL: "rig-operator:9000",
			},
		},
		Repository: Repository{
			Store:  "postgres",
			Secret: "",
		},
		Email: Email{
			From: "",
			Type: EmailTypeNoEmail,
		},
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
