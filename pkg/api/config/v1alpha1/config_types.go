package v1alpha1

import (
	"github.com/rigdev/rig/pkg/ptr"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// OperatorConfig is the Schema for the configs API
// +kubebuilder:object:root=true
type OperatorConfig struct {
	metav1.TypeMeta `json:",inline"`

	// WebhooksEnabled sets wether or not webhooks should be enabled. When
	// enabled a certificate should be mounted at the webhook server
	// certificate path. Defaults to true if omitted.
	WebhooksEnabled *bool `json:"webhooksEnabled,omitempty"`

	// DevModeEnabled enables verbose logs and changes the logging format to be
	// more human readable.
	DevModeEnabled bool `json:"devModeEnabled,omitempty"`

	// LeaderElectionEnabled enables leader election when running multiple
	// instances of the operator.
	LeaderElectionEnabled *bool `json:"leaderElectionEnabled,omitempty"`

	// Certmanager holds configuration for how the operator should create
	// certificates for ingress resources.
	Certmanager *CertManagerConfig `json:"certManager,omitempty"`

	// Service holds the configuration for service resources created by the
	// operator.
	Service ServiceConfig `json:"service,omitempty"`

	// Ingress holds the configuration for ingress resources created by the
	// operator.
	Ingress IngressConfig `json:"ingress,omitempty"`

	// PrometheusServiceMonitor defines if Rig should spawn a Prometheus ServiceMonitor per capsule
	// for use with a Prometheus Operator stack.
	PrometheusServiceMonitor *PrometheusServiceMonitor `json:"prometheusServiceMonitor,omitempty"`
}

type PrometheusServiceMonitor struct {
	// Path is the path which Prometheus should query on ports. Defaults to /metrics if not set.
	Path string `json:"path,omitempty"`
	// PortName is the name of the port which Prometheus will query metrics on
	PortName string `json:"portName"`
}

type CertManagerConfig struct {
	// ClusterIssuer to use for issueing ingress certificates
	ClusterIssuer string `json:"clusterIssuer"`

	// CreateCertificateResources specifies wether to create Certificate
	// resources. If this is not enabled we will use ingress annotations. This
	// is handy in environments where the ingress-shim isn't enabled.
	CreateCertificateResources bool `json:"createCertificateResources,omitempty"`
}

type ServiceConfig struct {
	// Type of the service to generate. By default, services are of type ClusterIP.
	// Valid values are ClusterIP, NodePort.
	Type corev1.ServiceType `json:"type"`
}

type IngressConfig struct {
	// Annotations for all ingress resources created.
	Annotations map[string]string `json:"annotations"`

	// ClassName specifies the default ingress class to use for all ingress
	// resources created.
	ClassName string `json:"className"`

	// PathType defines how ingress paths should be interpreted.
	// Allowed values: Exact, Prefix, ImplementationSpecific
	PathType networkingv1.PathType `json:"pathType"`

	// SkipTLS disables TLS configuration for ingress resources. This is useful
	// if a 3rd-party component is handling the HTTPS TLS termination and certificates.
	SkipTLS bool `json:"skip_tls"`
}

func (c *OperatorConfig) Default() {
	if c.WebhooksEnabled == nil {
		c.WebhooksEnabled = ptr.New(true)
	}
	if c.LeaderElectionEnabled == nil {
		c.LeaderElectionEnabled = ptr.New(true)
	}
	if c.Service.Type == "" {
		c.Service.Type = corev1.ServiceTypeClusterIP
	}
	if c.Ingress.Annotations == nil {
		c.Ingress.Annotations = map[string]string{}
	}
	if c.Ingress.PathType == "" {
		c.Ingress.PathType = networkingv1.PathTypeExact
	}
}

func init() {
	SchemeBuilder.Register(&OperatorConfig{})
	SchemeBuilder.Register(&PlatformConfig{})
}

// OperatorConfig is the Schema for the configs API
// +kubebuilder:object:root=true
type PlatformConfig struct {
	metav1.TypeMeta `json:",inline"`

	// Port sets the port the platform should listen on
	Port int `json:"port,omitempty"`

	// PublicUrl sets the public url for the platform. This is used for
	// generating urls for the platform when using oauth2.
	PublicURL string `json:"publicUrl,omitempty"`

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

	// Loggin holds information about the granularity of logging
	Logging Logging `json:"logging,omitempty"`

	// Clusters the platform has access to.
	Clusters []Cluster `json:"clusters,omitempty"`

	// Environments of the platform. Each environment is backed by a cluster (allowing multi-tenant setups).
	Environments []Environment `json:"environments,omitempty"`
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
	SSO *SSO `json:"sso,omitempty"`
}

// SSO specifies single sign on configuration.
type SSO struct {
	// OIDCProviders specifies enabled OIDCProviders which can be used for
	// login.
	OIDCProviders []OIDCProvider
}

// OIDCProvider specifies an OIDC provider.
type OIDCProvider struct {
	// Name is the name of the OIDC provider. This will be shown in the login
	// window.
	Name string `json:"name"`

	// IssuerURL is the URL for the OIDC issuer endpoint.
	IssuerURL string `json:"issuerURL"`

	// ClientID is the OAuth client ID.
	ClientID string `json:"clientId"`

	// ClientSecret is the OAuth client secret.
	ClientSecret string `json:"clientSecret"`

	// AllowedDomains is a list of email domains to allow. If left empty any
	// successful authentication on the provider is allowed.
	AllowedDomains []string `json:"allowedDomains,omitempty"`

	// GroupsClaim is the path to a claim in the JWT containing a string or
	// list of strings of group names.
	GroupsClaim string `json:"groupsClaim,omitempty"`

	// DisableJITGroups disables creation of groups found through OIDC in rig.
	DisableJITGroups *bool `json:"disableJITGroups,omitempty"`

	// GroupMapping is a mapping from OIDC provided group names to group names
	// used in rig. If an OIDC provided group name is not provided in this
	// mapping we will use the OIDC provided groupname in rig.
	GroupMapping map[string]string `json:"groupMapping,omitempty"`
}

// Client holds various client configuration
type Client struct {
	// Postgres holds configuration for the postgres client.
	Postgres ClientPostgres `json:"postgres,omitempty"`

	// Mongo holds configuration for the Mongo client.
	Mongo ClientMongo `json:"mongo,omitempty"`

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

// ClientMongo specifies the configuration for the mongo client.
type ClientMongo struct {
	// User is the database user used when connecting to the mongodb server.
	User string `json:"user,omitempty"`

	// Password is used when connecting to the mongodb server.
	Password string `json:"password,omitempty"`

	// Host of the mongo server. This is both the host and port.
	Host string `json:"host,omitempty"`
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
	// Store is what database will be used can be either postgres or mongodb.
	Store string `json:"store,omitempty"`

	// Secret is a secret key used for encrypting sensitive data before saving
	// it in the database.
	Secret string `json:"secret,omitempty"`
}

// Cluster specifies cluster configuration
type Cluster struct {
	// Name of the cluster. The name is used as a reference for the cluster through the documentation
	// and API endpoints.
	Name string `json:"name,omitempty"`

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
	// Name of the environment.
	Name string `json:"name,omitempty"`

	// Cluster name the environment is hosted in.
	Cluster string `json:"cluster,omitempty"`

	// NamespaceTemplate is used to generate the namespace name when configuring resources.
	// Default is to set the namespace equal to the project name.
	NamespaceTemplate string `json:"namespace_template,omitempty"`
}

// ClusterGit specifies configuration for git integration. This can be used to
// tie rig into a gitops setup.
type ClusterGit struct {
	// URL is the git repository URL.
	URL string `json:"url,omitempty"`

	// Branch to commit changes to.
	Branch string `json:"branch,omitempty"`

	// PathPrefix path to commit to in git repository.
	PathPrefix string `json:"pathPrefix,omitempty"`

	// Credentials to use when connecting to git.
	Credentials GitCredentials `json:"credentials,omitempty"`

	// Author used when creating commits.
	Author GitAuthor `json:"author,omitempty"`

	// Templates used for commit messages.
	Templates GitTemplates `json:"templates,omitempty"`
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
			Mongo: ClientMongo{
				Host: "",
			},
			Docker: ClientDocker{
				Host: "",
			},
			Mailjet: ClientMailjet{
				APIKey:    "",
				SecretKey: "",
			},
			Operator: ClientOperator{
				BaseURL: "",
			},
		},
		Repository: Repository{
			Store:  "postgres",
			Secret: "",
		},
		Cluster: Cluster{
			Type: ClusterTypeDocker,
			Git: ClusterGit{
				Branch:     "main",
				PathPrefix: `apps/{{ .Project.Name }}/{{ .Capsule.Name }}/`,
				Templates: GitTemplates{
					Rollout: `Rig Platform rollout #{{ .Rollout.ID }} of {{ .Capsule.Name }}

Rollout initiated by {{ .Initiator.Name }} at {{ .Rollout.CreatedAt }}.
`,
					Delete: `Rig Platform delete of {{ .Capsule.Name }}

Capsule deleted by {{ .Initiator.Name }}.
`,
				},
				Author: GitAuthor{
					Name:  "rig-platform-change-roller",
					Email: "roll@rig.dev",
				},
			},
		},
		Email: Email{
			From: "",
			Type: EmailTypeNoEmail,
		},
		Environments: []Environment{
			{
				Name:              "default",
				Cluster:           "default",
				NamespaceTemplate: "{{ .Project.Name }}",
			},
		},
	}

	cfg.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "config.rig.dev",
		Version: "v1alpha1",
		Kind:    "PlatformConfig",
	})

	return cfg
}
